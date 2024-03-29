package timelines

import (
	"context"
	"errors"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/google/uuid"
	"yatc/internal"
	statuses "yatc/status/pkg"
	timelines "yatc/timeline/pkg"
	"yatc/user/pkg/followers"
)

type Service struct {
	repo            Repository
	followerService followers.Service
	client          dapr.Client
	config          internal.PubSubConfig
}

func NewTimelineService(repo Repository, followerService followers.Service, client dapr.Client, config internal.PubSubConfig) *Service {
	return &Service{repo, followerService, client, config}
}

func (timelineService *Service) GetTimeline(ctx context.Context, userId uuid.UUID) (timelines.Timeline, error) {
	return timelineService.repo.Get(userId)
}

func (timelineService *Service) UpdateTimelines(ctx context.Context, userId uuid.UUID, status statuses.Status) error {
	allFollowers, err := timelineService.followerService.GetFollowers(ctx, userId)
	if err != nil {
		return err
	}
	for _, follower := range allFollowers {
		timeline, err := timelineService.GetTimeline(context.Background(), follower.Id)
		if err != nil {
			if errors.Is(err, internal.NotFoundError(follower.Id)) {
				timeline = timelines.Timeline{
					UserId:   follower.Id,
					Statuses: []statuses.Status{status},
				}
			}
		} else {
			timeline.Statuses = append(timeline.Statuses, status)
		}

		_, err = timelineService.repo.Save(timeline)
		if err != nil {
			return err
		}
	}

	err = timelineService.client.PublishEvent(context.Background(), timelineService.config.Name, "timeline", "")
	if err != nil {
		return err
	}

	return nil
}
