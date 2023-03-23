package followers

import (
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

func (service *Service) GetFollowers(userId uuid.UUID) ([]users.User, error) {
	user, err := service.repo.Get(userId)
	if err != nil {
		return nil, err
	}
	followers := make([]users.User, 0)
	for followerId := range user.Followers {
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

func (service *Service) GetFollowees(userId uuid.UUID) ([]users.User, error) {
	user, err := service.repo.Get(userId)
	if err != nil {
		return nil, err
	}
	followees := make([]users.User, 0)
	for followerId := range user.Followees {
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

func (service *Service) FollowUser(userToFollowId uuid.UUID, userWhichFollowsId uuid.UUID) (users.User, error) {
	if userWhichFollowsId == userToFollowId {
		return users.User{}, SelfFollowError
	}

	// Add follower to user followers
	userToFollow, err := service.repo.Get(userToFollowId)
	if err != nil {
		return users.User{}, err
	}
	userToFollow.Followers[userWhichFollowsId] = struct{}{}

	// Add user to followee of follower
	userWhichFollows, err := service.repo.Get(userWhichFollowsId)
	if err != nil {
		return users.User{}, err
	}
	userWhichFollows.Followees[userToFollowId] = struct{}{}

	userToFollow, err = service.repo.Save(userToFollow)
	_, err = service.repo.Save(userWhichFollows)

	return userToFollow, err
}

func (service *Service) UnfollowUser(userToUnfollowId uuid.UUID, userWhichUnfollowsId uuid.UUID) error {
	if userWhichUnfollowsId == userToUnfollowId {
		return SelfFollowError
	}

	// Delete follower from user followers
	userToUnfollow, err := service.repo.Get(userToUnfollowId)
	if err != nil {
		return err
	}
	if _, exists := userToUnfollow.Followers[userWhichUnfollowsId]; !exists {
		return internal.NotFoundError(userWhichUnfollowsId)
	}
	delete(userToUnfollow.Followers, userWhichUnfollowsId)

	// Delete user from followees of follower
	userWhichUnfollows, err := service.repo.Get(userWhichUnfollowsId)
	if err != nil {
		return err
	}
	if _, exists := userWhichUnfollows.Followees[userToUnfollowId]; !exists {
		return internal.NotFoundError(userToUnfollowId)
	}
	delete(userWhichUnfollows.Followees, userToUnfollowId)

	userToUnfollow, err = service.repo.Save(userToUnfollow)
	_, err = service.repo.Save(userWhichUnfollows)

	return err
}
