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
	UpdateTimelines(userId uuid.UUID, status statuses.Status) error
}
