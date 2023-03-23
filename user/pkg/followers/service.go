package followers

import (
	"github.com/google/uuid"
	"yatc/user/pkg/users"
)

type Service interface {
	GetFollowers(userId uuid.UUID) ([]users.User, error)
	GetFollowees(userId uuid.UUID) ([]users.User, error)
	FollowUser(userToFollowId uuid.UUID, userWhichFollowsId uuid.UUID) (users.User, error)
	UnfollowUser(userToFollowId uuid.UUID, userWhichFollowsId uuid.UUID) error
}
