package followers

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"yatc/internal"
	iusers "yatc/user/internal/users"
	"yatc/user/pkg/users"
)

var SelfFollowError = errors.New("cant follow one self")

type Service struct {
	repo iusers.Repository
}

func NewFollowerService(repo iusers.Repository) *Service {
	return &Service{repo: repo}
}

func (service *Service) GetFollowers(ctx context.Context, userId uuid.UUID) ([]users.User, error) {
	user, err := service.repo.Get(userId)
	if err != nil {
		return nil, err
	}
	followers := make([]users.User, 0)
	for _, followerId := range user.Followers.ToArray() {
		follower, err := service.repo.Get(followerId)
		if err != nil {
			if errors.Is(err, internal.NotFoundError(userId)) {
				return nil, fmt.Errorf("user not found while populating followers array. error: %w", err)
			} else {
				return nil, err
			}
		}

		followers = append(followers, follower)
	}
	return followers, err
}

func (service *Service) GetFollowees(ctx context.Context, userId uuid.UUID) ([]users.User, error) {
	user, err := service.repo.Get(userId)
	if err != nil {
		return nil, err
	}
	followees := make([]users.User, 0)
	for _, followerId := range user.Followees.ToArray() {
		followee, err := service.repo.Get(followerId)
		if err != nil {
			if errors.Is(err, internal.NotFoundError(userId)) {
				return nil, fmt.Errorf("user not found while populating followees array. error: %w", err)
			} else {
				return nil, err
			}
		}

		followees = append(followees, followee)
	}
	return followees, err
}

func (service *Service) FollowUser(ctx context.Context, userToFollowId uuid.UUID, userWhichFollowsId uuid.UUID) (users.User, error) {
	if userWhichFollowsId == userToFollowId {
		return users.User{}, SelfFollowError
	}

	// Add follower to user followers
	userToFollow, err := service.repo.Get(userToFollowId)
	if err != nil {
		return users.User{}, err
	}
	userToFollow.Followers.Add(userWhichFollowsId)

	// Add user to followee of follower
	userWhichFollows, err := service.repo.Get(userWhichFollowsId)
	if err != nil {
		return users.User{}, err
	}
	userWhichFollows.Followees.Add(userToFollowId)

	userToFollow, err = service.repo.Save(userToFollow)
	if err != nil {
		return users.User{}, err
	}

	_, err = service.repo.Save(userWhichFollows)
	if err != nil {
		return users.User{}, err
	}

	return userToFollow, nil
}

func (service *Service) UnfollowUser(ctx context.Context, userToUnfollowId uuid.UUID, userWhichUnfollowsId uuid.UUID) error {
	if userWhichUnfollowsId == userToUnfollowId {
		return SelfFollowError
	}

	// Delete follower from user followers
	userToUnfollow, err := service.repo.Get(userToUnfollowId)
	if err != nil {
		return err
	}
	if !userToUnfollow.Followers.Has(userWhichUnfollowsId) {
		return internal.NotFoundError(userWhichUnfollowsId)
	}
	userToUnfollow.Followers.Remove(userWhichUnfollowsId)

	// Delete user from followees of follower
	userWhichUnfollows, err := service.repo.Get(userWhichUnfollowsId)
	if err != nil {
		return err
	}
	if !userWhichUnfollows.Followees.Has(userToUnfollowId) {
		return internal.NotFoundError(userToUnfollowId)
	}
	userWhichUnfollows.Followees.Remove(userToUnfollowId)

	userToUnfollow, err = service.repo.Save(userToUnfollow)
	if err != nil {
		return err
	}
	_, err = service.repo.Save(userWhichUnfollows)
	if err != nil {
		return err
	}

	return nil
}
