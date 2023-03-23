package timelines

import (
	"github.com/google/uuid"
	"yatc/internal"
	timelines "yatc/timeline/pkg"
)

type Repository interface {
	Get(userId uuid.UUID) (timelines.Timeline, error)
	Save(timeline timelines.Timeline) (timelines.Timeline, error)
}

type InMemoryRepo struct {
	Timelines map[uuid.UUID]timelines.Timeline
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{Timelines: map[uuid.UUID]timelines.Timeline{}}
}

func (repo InMemoryRepo) Get(userId uuid.UUID) (timelines.Timeline, error) {
	allTimelines, ok := repo.Timelines[userId]
	if !ok {
		return timelines.Timeline{}, internal.NotFoundError(userId)
	}
	return allTimelines, nil
}

func (repo InMemoryRepo) Save(timeline timelines.Timeline) (timelines.Timeline, error) {
	repo.Timelines[timeline.UserId] = timeline
	return timeline, nil
}
