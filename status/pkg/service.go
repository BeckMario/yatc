package statuses

import (
	"github.com/google/uuid"
)

type Status struct {
	Id      uuid.UUID `db:"id"`
	Content string    `db:"content"`
	UserId  uuid.UUID `db:"user_id"`
}

type Service interface {
	GetStatuses() ([]Status, error)
	GetStatus(statusId uuid.UUID) (Status, error)
	CreateStatus(status Status) (Status, error)
	DeleteStatus(statusId uuid.UUID) (Status, error)
}
