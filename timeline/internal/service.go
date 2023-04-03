package timelines

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"yatc/internal"
	statuses "yatc/status/pkg"
	timelines "yatc/timeline/pkg"
	"yatc/user/pkg/followers"
)

type Service struct {
	repo            Repository
	followerService followers.Service
}

func NewTimelineService(repo Repository, followerService followers.Service) *Service {
	return &Service{repo, followerService}
}

func (timelineService *Service) GetTimeline(userId uuid.UUID) (timelines.Timeline, error) {
	return timelineService.repo.Get(userId)
}

func (timelineService *Service) UpdateTimelines(ctx context.Context, userId uuid.UUID, status statuses.Status) error {
	allFollowers, err := timelineService.followerService.GetFollowers(ctx, userId)
	if err != nil {
		return err
	}
	for _, follower := range allFollowers {
		timeline, err := timelineService.GetTimeline(follower.Id)
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

	return nil
}
