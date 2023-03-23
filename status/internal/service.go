package statuses

import (
	"github.com/google/uuid"
	"yatc/status/pkg"
)

type Service struct {
	repo      Repository
	publisher Publisher
}

func NewStatusService(repo Repository, publisher Publisher) *Service {
	return &Service{repo: repo, publisher: publisher}
}

func (statusService *Service) GetStatuses() ([]statuses.Status, error) {
	return statusService.repo.List()
}

func (statusService *Service) GetStatus(statusId uuid.UUID) (statuses.Status, error) {
	return statusService.repo.Get(statusId)
}

func (statusService *Service) CreateStatus(status statuses.Status) (statuses.Status, error) {
	createdStatus, err := statusService.repo.Create(status)
	if err != nil {
		return statuses.Status{}, err
	}

	err = statusService.publisher.Publish(status)
	if err != nil {
		return statuses.Status{}, err
	}

	return createdStatus, nil
}

func (statusService *Service) DeleteStatus(statusId uuid.UUID) (statuses.Status, error) {
	return statusService.repo.Delete(statusId)
}
