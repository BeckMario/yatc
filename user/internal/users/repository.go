package users

import (
	"context"
	"encoding/json"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/google/uuid"
	"yatc/internal"
	"yatc/user/pkg/users"
)

type Repository interface {
	List() ([]users.User, error)
	Get(userId uuid.UUID) (users.User, error)
	Delete(userId uuid.UUID) (users.User, error)
	Save(user users.User) (users.User, error)
}

type DaprStateStoreRepo struct {
	dapr   dapr.Client
	config internal.StateStoreConfig
}

func NewDaprRepo(client dapr.Client, config internal.StateStoreConfig) Repository {
	return &DaprStateStoreRepo{client, config}
}

func (repo *DaprStateStoreRepo) List() ([]users.User, error) {
	item, err := repo.dapr.GetState(context.Background(), repo.config.Name, "keys", nil)
	if err != nil {
		return nil, err
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

	allUsers := make([]users.User, len(states))
	for i, state := range states {
		var user users.User
		err = json.Unmarshal(state.Value, &user)
		if err != nil {
			return nil, err
		}
		allUsers[i] = user
	}

	return allUsers, nil
}

func (repo *DaprStateStoreRepo) Get(userId uuid.UUID) (users.User, error) {
	userItem, err := repo.dapr.GetState(context.Background(), repo.config.Name, userId.String(), nil)
	if err != nil {
		return users.User{}, err
	}
	if userItem.Value == nil {
		return users.User{}, internal.NotFoundError(userId)
	}

	var user users.User
	err = json.Unmarshal(userItem.Value, &user)
	if err != nil {
		return users.User{}, err
	}

	return user, nil
}

func (repo *DaprStateStoreRepo) Delete(userId uuid.UUID) (users.User, error) {
	ctx := context.Background()
	userItem, err := repo.dapr.GetState(ctx, repo.config.Name, userId.String(), nil)
	if err != nil {
		return users.User{}, err
	}

	if userItem.Value == nil {
		return users.User{}, internal.NotFoundError(userId)
	}

	err = repo.dapr.DeleteState(ctx, repo.config.Name, userId.String(), nil)
	if err != nil {
		return users.User{}, err
	}

	var user users.User
	err = json.Unmarshal(userItem.Value, &user)
	if err != nil {
		return users.User{}, err
	}

	return user, nil
}

func (repo *DaprStateStoreRepo) Save(user users.User) (users.User, error) {
	ctx := context.Background()
	userJson, err := json.Marshal(user)
	if err != nil {
		return users.User{}, err
	}

	item, err := repo.dapr.GetState(ctx, repo.config.Name, "keys", nil)
	if err != nil {
		return users.User{}, err
	}

	if item.Value == nil {
		keySet := internal.Ptr(internal.NewSet[uuid.UUID]())
		bytes, err := json.Marshal(keySet)
		if err != nil {
			return users.User{}, err
		}
		item.Value = bytes
	}

	var keySet *internal.Set[uuid.UUID]
	err = json.Unmarshal(item.Value, &keySet)
	if err != nil {
		return users.User{}, err
	}

	keySet.Add(user.Id)

	keysJson, err := json.Marshal(keySet)
	if err != nil {
		return users.User{}, err
	}

	saveUserOp := dapr.StateOperation{
		Type: dapr.StateOperationTypeUpsert,
		Item: &dapr.SetStateItem{
			Key:   user.Id.String(),
			Value: userJson,
		},
	}

	saveKeyOp := dapr.StateOperation{
		Type: dapr.StateOperationTypeUpsert,
		Item: &dapr.SetStateItem{
			Key:   "keys",
			Value: keysJson,
		},
	}

	err = repo.dapr.ExecuteStateTransaction(ctx, repo.config.Name, nil, []*dapr.StateOperation{&saveUserOp, &saveKeyOp})
	if err != nil {
		return users.User{}, err
	}

	return user, nil
}

type InMemoryRepo struct {
	Users map[uuid.UUID]users.User
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{Users: map[uuid.UUID]users.User{}}
}

func (repo InMemoryRepo) List() ([]users.User, error) {
	values := make([]users.User, 0, len(repo.Users))

	for _, value := range repo.Users {
		values = append(values, value)
	}

	return values, nil
}

func (repo InMemoryRepo) Get(userId uuid.UUID) (users.User, error) {
	user, ok := repo.Users[userId]
	if !ok {
		return users.User{}, internal.NotFoundError(userId)
	}
	return user, nil
}

func (repo InMemoryRepo) Delete(userId uuid.UUID) (users.User, error) {
	user, exists := repo.Users[userId]
	if !exists {
		return users.User{}, internal.NotFoundError(userId)
	}
	delete(repo.Users, userId)
	return user, nil
}

func (repo InMemoryRepo) Save(user users.User) (users.User, error) {
	repo.Users[user.Id] = user
	return user, nil
}
