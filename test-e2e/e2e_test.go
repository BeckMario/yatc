package test_e2e

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
	"yatc/internal"
	"yatc/login/pkg"
	statuses "yatc/status/pkg"
	timelines "yatc/timeline/pkg"
	"yatc/user/pkg/followers"
	"yatc/user/pkg/users"
)

func TestCreateStatus(t *testing.T) {
	logger, sync := internal.NewZapLogger()
	defer sync(logger)

	config := internal.NewConfig("config.yaml", logger)

	userClient := users.NewUserClient(config.Dapr)
	followerClient := followers.NewFollowerClient(config.Dapr)
	statusClient := statuses.NewStatusClient(config.Dapr)
	timelineClient := timelines.NewTimelineClient(config.Dapr)
	loginClient := login.NewLoginClient(config.Dapr)

	// Create User 1
	user1, err := userClient.CreateUser(users.User{
		Name: "Luke",
	})
	assert.NoError(t, err)

	// Create User 2
	user2, err := userClient.CreateUser(users.User{
		Name: "Grogu",
	})
	assert.NoError(t, err)

	jwtUser1, err := loginClient.Login(user1.Name, user1.Id)
	assert.NoError(t, err)

	jwtUser2, err := loginClient.Login(user2.Name, user2.Id)
	assert.NoError(t, err)

	// User 1 follows User 2
	_, err = followerClient.FollowUser(context.WithValue(context.Background(), internal.ContextKeyAuthorization, jwtUser1), user2.Id, user1.Id)
	assert.NoError(t, err)

	// User 2 Creates Status
	status := statuses.Status{
		Id:      uuid.UUID{},
		Content: "New Status",
		UserId:  user2.Id,
	}
	createdStatus, err := statusClient.CreateStatus(context.WithValue(context.Background(), internal.ContextKeyAuthorization, jwtUser2), status)
	assert.NoError(t, err)

	// Wait for Creation of Timeline, because Async Communication with Pub/Sub between Status and Timeline Service is happening
	// TODO: For more consistency i could do the test with a dapr sidecar and subscribe the status topic
	time.Sleep(50 * time.Millisecond)

	// User 1 Get Timeline
	timeline, err := timelineClient.GetTimeline(context.WithValue(context.Background(), internal.ContextKeyAuthorization, jwtUser1), user1.Id)
	assert.NoError(t, err)

	// User 1 Timeline Contains Status of User 2
	found := false
	for _, status := range timeline.Statuses {
		if status.Id == createdStatus.Id {
			found = true
		}
	}

	assert.True(t, found, "timeline contains status")
}
