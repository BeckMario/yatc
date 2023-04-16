package statuses

import (
	"context"
	"encoding/json"
	"errors"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"yatc/internal"
	"yatc/status/pkg"
)

type Repository interface {
	List() ([]statuses.Status, error)
	Get(statusId uuid.UUID) (statuses.Status, error)
	Delete(statusId uuid.UUID) (statuses.Status, error)
	Create(status statuses.Status) (statuses.Status, error)
}

type InMemoryRepo struct {
	Statuses map[uuid.UUID]statuses.Status
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{Statuses: map[uuid.UUID]statuses.Status{}}
}

func (repo InMemoryRepo) List() ([]statuses.Status, error) {
	values := make([]statuses.Status, 0, len(repo.Statuses))

	for _, value := range repo.Statuses {
		values = append(values, value)
	}

	return values, nil
}

func (repo InMemoryRepo) Get(statusId uuid.UUID) (statuses.Status, error) {
	status, ok := repo.Statuses[statusId]
	if !ok {
		return statuses.Status{}, internal.NotFoundError(statusId)
	}
	return status, nil
}

func (repo InMemoryRepo) Delete(statusId uuid.UUID) (statuses.Status, error) {
	status, exists := repo.Statuses[statusId]
	if !exists {
		return statuses.Status{}, internal.NotFoundError(statusId)
	}
	delete(repo.Statuses, statusId)
	return status, nil
}

func (repo InMemoryRepo) Create(status statuses.Status) (statuses.Status, error) {
	_, exists := repo.Statuses[status.Id]
	if exists {
		return statuses.Status{}, errors.New("duplicated status")
	}
	repo.Statuses[status.Id] = status
	return status, nil
}

type PostgresRepo struct {
	db *sqlx.DB
}

func NewPostgresRepo(db *sqlx.DB) Repository {
	return &PostgresRepo{db}
}

func (r *PostgresRepo) List() ([]statuses.Status, error) {
	var allStatuses []statuses.Status
	err := r.db.Select(&allStatuses, "SELECT * FROM statuses")
	if err != nil {
		return nil, err
	}
	return allStatuses, nil
}

func (r *PostgresRepo) Get(statusId uuid.UUID) (statuses.Status, error) {
	status := statuses.Status{}
	err := r.db.Get(&status, "SELECT * FROM statuses WHERE id=$1", statusId)
	if err != nil {
		return statuses.Status{}, internal.NotFoundError(statusId)
	}
	return status, nil
}

func (r *PostgresRepo) Delete(statusId uuid.UUID) (statuses.Status, error) {
	status, err := r.Get(statusId)
	if err != nil {
		return statuses.Status{}, err
	}
	_, err = r.db.Exec("DELETE FROM statuses WHERE id=$1", statusId)
	if err != nil {
		return statuses.Status{}, err
	}
	return status, nil
}

func (r *PostgresRepo) Create(status statuses.Status) (statuses.Status, error) {
	_, err := r.db.Exec("INSERT INTO statuses (id, content, user_id) VALUES ($1, $2, $3)", status.Id, status.Content, status.UserId)
	if err != nil {
		return statuses.Status{}, err
	}
	return status, nil
}

type DaprStateStoreRepo struct {
	dapr   dapr.Client
	config internal.StateStoreConfig
}

func NewDaprStateStore(client dapr.Client, config internal.StateStoreConfig) *DaprStateStoreRepo {
	return &DaprStateStoreRepo{client, config}
}

func (repo *DaprStateStoreRepo) List() ([]statuses.Status, error) {
	item, err := repo.dapr.GetState(context.Background(), repo.config.Name, "keys", nil)
	if err != nil {
		return nil, err
	}

	if item.Value == nil {
		return make([]statuses.Status, 0), nil
	}

	listOfKeys := make([]string, 0)
	err = json.Unmarshal(item.Value, &listOfKeys)
	if err != nil {
		return nil, err
	}

	states, err := repo.dapr.GetBulkState(context.Background(), repo.config.Name, listOfKeys, nil, 1)
	if err != nil {
		return nil, err
	}

	allStatuses := make([]statuses.Status, len(states))
	for i, state := range states {
		var status statuses.Status
		err = json.Unmarshal(state.Value, &status)
		if err != nil {
			return nil, err
		}
		allStatuses[i] = status
	}

	return allStatuses, nil
}

func (repo *DaprStateStoreRepo) Get(statusId uuid.UUID) (statuses.Status, error) {
	statusItem, err := repo.dapr.GetState(context.Background(), repo.config.Name, statusId.String(), nil)
	if err != nil {
		return statuses.Status{}, err
	}
	if statusItem.Value == nil {
		return statuses.Status{}, internal.NotFoundError(statusId)
	}

	var status statuses.Status
	err = json.Unmarshal(statusItem.Value, &status)
	if err != nil {
		return statuses.Status{}, err
	}

	return status, nil
}

func (repo *DaprStateStoreRepo) Delete(statusId uuid.UUID) (statuses.Status, error) {
	ctx := context.Background()
	statusItem, err := repo.dapr.GetState(ctx, repo.config.Name, statusId.String(), nil)
	if err != nil {
		return statuses.Status{}, err
	}

	if statusItem.Value == nil {
		return statuses.Status{}, internal.NotFoundError(statusId)
	}

	err = repo.dapr.DeleteState(ctx, repo.config.Name, statusId.String(), nil)
	if err != nil {
		return statuses.Status{}, err
	}

	var status statuses.Status
	err = json.Unmarshal(statusItem.Value, &status)
	if err != nil {
		return statuses.Status{}, err
	}

	return status, nil
}

func (repo *DaprStateStoreRepo) Create(status statuses.Status) (statuses.Status, error) {
	ctx := context.Background()
	statusJson, err := json.Marshal(status)
	if err != nil {
		return statuses.Status{}, err
	}

	item, err := repo.dapr.GetState(ctx, repo.config.Name, "keys", nil)
	if err != nil {
		return statuses.Status{}, err
	}

	if item.Value == nil {
		keySet := internal.Ptr(internal.NewSet[uuid.UUID]())
		bytes, err := json.Marshal(keySet)
		if err != nil {
			return statuses.Status{}, err
		}
		item.Value = bytes
	}

	var keySet *internal.Set[uuid.UUID]
	err = json.Unmarshal(item.Value, &keySet)
	if err != nil {
		return statuses.Status{}, err
	}

	keySet.Add(status.Id)

	keysJson, err := json.Marshal(keySet)
	if err != nil {
		return statuses.Status{}, err
	}

	saveStatusOp := dapr.StateOperation{
		Type: dapr.StateOperationTypeUpsert,
		Item: &dapr.SetStateItem{
			Key:   status.Id.String(),
			Value: statusJson,
		},
	}

	saveKeyOp := dapr.StateOperation{
		Type: dapr.StateOperationTypeUpsert,
		Item: &dapr.SetStateItem{
			Key:   "keys",
			Value: keysJson,
		},
	}

	err = repo.dapr.ExecuteStateTransaction(ctx, repo.config.Name, nil, []*dapr.StateOperation{&saveStatusOp, &saveKeyOp})
	if err != nil {
		return statuses.Status{}, err
	}

	return status, nil
}
