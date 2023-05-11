package statuses

import (
	"context"
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

func (statusService *Service) GetStatuses(userId uuid.UUID) ([]statuses.Status, error) {
	list, err := statusService.repo.List()
	if err != nil {
		return nil, err
	}

	foundStatuses := make([]statuses.Status, 0)
	for _, status := range list {
		if status.UserId == userId {
			foundStatuses = append(foundStatuses, status)
		}
	}

	return foundStatuses, nil
}

func (statusService *Service) GetStatus(statusId uuid.UUID) (statuses.Status, error) {
	return statusService.repo.Get(statusId)
}

func (statusService *Service) CreateStatus(ctx context.Context, status statuses.Status) (statuses.Status, error) {
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
