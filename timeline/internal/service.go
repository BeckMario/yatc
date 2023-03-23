package timelines

import (
	"errors"
	"fmt"
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

func (timelineService *Service) GetTimeline(userId uuid.UUID) (timelines.Timeline, error) {
	return timelineService.repo.Get(userId)
}

func (timelineService *Service) UpdateTimeline(userId uuid.UUID, status statuses.Status) (timelines.Timeline, error) {
	fmt.Println("Getting Followers")
	allFollowers, err := timelineService.followerService.GetFollowers(userId)
	if err != nil {
		return timelines.Timeline{}, err
	}
	fmt.Println(allFollowers)
	for _, follower := range allFollowers {
		fmt.Println("Insede for")
		timeline, err := timelineService.GetTimeline(follower.Id)
		if err != nil {
			if errors.Is(err, internal.NotFoundError(follower.Id)) {
				fmt.Println("No timeline found, therefore create one")
				timeline = timelines.Timeline{
					UserId:   follower.Id,
					Statuses: []statuses.Status{status},
				}
			}
		} else {
			fmt.Println("timeline found, therefore add status")
			timeline.Statuses = append(timeline.Statuses, status)
		}

		savedTimeline, err := timelineService.repo.Save(timeline)
		if err != nil {
			return timelines.Timeline{}, err
		}
		fmt.Println(savedTimeline)
	}

	return timelines.Timeline{}, nil
}

func NewTweetService(repo Repository, followerService followers.Service) *Service {
	return &Service{repo, followerService}
}
