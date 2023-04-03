package followers

import (
	"context"
	"github.com/google/uuid"
	"yatc/user/pkg/users"
)

type Service interface {
	GetFollowers(ctx context.Context, userId uuid.UUID) ([]users.User, error)
	GetFollowees(ctx context.Context, userId uuid.UUID) ([]users.User, error)
	FollowUser(ctx context.Context, userToFollowId uuid.UUID, userWhichFollowsId uuid.UUID) (users.User, error)
	UnfollowUser(ctx context.Context, userToFollowId uuid.UUID, userWhichFollowsId uuid.UUID) error
}
