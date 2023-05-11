package statuses

import (
	"context"
	"github.com/google/uuid"
)

type Status struct {
	Id       uuid.UUID `db:"id"`
	Content  string    `db:"content"`
	UserId   uuid.UUID `db:"user_id"`
	MediaIds []uuid.UUID
}

type Service interface {
	GetStatuses(userId uuid.UUID) ([]Status, error)
	GetStatus(statusId uuid.UUID) (Status, error)
	CreateStatus(ctx context.Context, status Status) (Status, error)
	DeleteStatus(statusId uuid.UUID) (Status, error)
}
