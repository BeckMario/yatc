package statuses

import "github.com/google/uuid"

type Status struct {
	Id      uuid.UUID
	Content string
	UserId  uuid.UUID
}

type Service interface {
	GetStatuses() ([]Status, error)
	GetStatus(statusId uuid.UUID) (Status, error)
	CreateStatus(status Status) (Status, error)
	DeleteStatus(statusId uuid.UUID) (Status, error)
}
