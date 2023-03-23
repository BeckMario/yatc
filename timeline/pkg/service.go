package timelines

import (
	"github.com/google/uuid"
	"yatc/status/pkg"
)

type Timeline struct {
	UserId   uuid.UUID
	Statuses []statuses.Status
}

type Service interface {
	GetTimeline(userId uuid.UUID) (Timeline, error)
	UpdateTimeline(userId uuid.UUID, status statuses.Status) (Timeline, error)
}
