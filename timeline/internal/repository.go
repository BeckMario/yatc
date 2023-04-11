package timelines

import (
	"context"
	"encoding/json"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/google/uuid"
	"yatc/internal"
	timelines "yatc/timeline/pkg"
)

type Repository interface {
	Get(userId uuid.UUID) (timelines.Timeline, error)
	Save(timeline timelines.Timeline) (timelines.Timeline, error)
}

type DaprStateStoreRepo struct {
	dapr   dapr.Client
	config internal.StateStoreConfig
}

func NewDaprRepo(client dapr.Client, config internal.StateStoreConfig) *DaprStateStoreRepo {
	return &DaprStateStoreRepo{client, config}
}

func (repo *DaprStateStoreRepo) Get(userId uuid.UUID) (timelines.Timeline, error) {
	item, err := repo.dapr.GetState(context.Background(), repo.config.Name, userId.String(), nil)
	if err != nil {
		return timelines.Timeline{}, err
	}
	if item.Value == nil {
		return timelines.Timeline{}, internal.NotFoundError(userId)
	}

	var timeline timelines.Timeline
	err = json.Unmarshal(item.Value, &timeline)
	if err != nil {
		return timelines.Timeline{}, err
	}

	return timeline, nil
}

func (repo *DaprStateStoreRepo) Save(timeline timelines.Timeline) (timelines.Timeline, error) {
	timelineJson, err := json.Marshal(timeline)
	if err != nil {
		return timelines.Timeline{}, err
	}

	err = repo.dapr.SaveState(context.Background(), repo.config.Name, timeline.UserId.String(), timelineJson, nil)
	if err != nil {
		return timelines.Timeline{}, err
	}
	return timeline, nil
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
