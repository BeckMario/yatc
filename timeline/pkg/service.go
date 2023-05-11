package timelines

import (
	"context"
	"github.com/google/uuid"
	"yatc/status/pkg"
)

type Timeline struct {
	UserId   uuid.UUID
	Statuses []statuses.Status
}

type Service interface {
	GetTimeline(ctx context.Context, userId uuid.UUID) (Timeline, error)
	UpdateTimelines(ctx context.Context, userId uuid.UUID, status statuses.Status) error
}
