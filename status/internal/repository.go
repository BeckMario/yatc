package statuses

import (
	"errors"
	"github.com/google/uuid"
	statuses "yatc/status/pkg"
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
