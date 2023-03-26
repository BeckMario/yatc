package statuses

import (
	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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
		return statuses.Status{}, errors.New("no status found")
	}
	return status, nil
}

func (repo InMemoryRepo) Delete(statusId uuid.UUID) (statuses.Status, error) {
	status, exists := repo.Statuses[statusId]
	if !exists {
		return statuses.Status{}, errors.New("no status found")
	}
	delete(repo.Statuses, statusId)
	return status, nil
}

func (repo InMemoryRepo) Create(status statuses.Status) (statuses.Status, error) {
	repo.Statuses[status.Id] = status
	return status, nil
}

type PostgreSQLRepo struct {
	db *sqlx.DB
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
		return statuses.Status{}, err
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
