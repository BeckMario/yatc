//go:build e2e
// +build e2e

package test_e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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
	logger := zap.NewNop()

	config := internal.NewConfig("config.yaml", logger)
	config.Dapr.AppId = "krakend-service"

	statusChan := make(chan statuses.Status, 1)
	timelineChan := make(chan bool, 1)
	setUpPubSub(t, config, statusChan, timelineChan)

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

	// Waiting for status pub sub
	select {
	case statusR := <-statusChan:
		assert.Equal(t, createdStatus.Id, statusR.Id, "status in pubsub is equal to created status")
	case <-time.After(5 * time.Second):
		assert.Fail(t, "5s timeout while waiting for pub/sub status event")
	}

	// Waiting for timeline pub sub
	select {
	case <-timelineChan:
		break
	case <-time.After(5 * time.Second):
		assert.Fail(t, "5s timeout while waiting for pub/sub timeline event")
	}

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

func setUpPubSub(t *testing.T, config *internal.Config, statusChan chan statuses.Status, timelineChan chan bool) {
	s := daprd.NewService(fmt.Sprintf(":%s", config.Port))
	subStatus := &common.Subscription{
		PubsubName: config.Dapr.PubSub.Name,
		Topic:      "status",
		Route:      "/pubsub/status",
	}
	subTimeline := &common.Subscription{
		PubsubName: config.Dapr.PubSub.Name,
		Topic:      "timeline",
		Route:      "/pubsub/timeline",
	}
	err := s.AddTopicEventHandler(subStatus, func(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
		var status statuses.Status
		data := e.RawData
		_ = json.Unmarshal(data, &status)
		statusChan <- status
		return false, nil
	})
	assert.NoError(t, err)
	err = s.AddTopicEventHandler(subTimeline, func(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
		timelineChan <- true
		return false, nil
	})
	assert.NoError(t, err)

	go func() {
		err := s.Start()
		assert.NoError(t, err)
	}()
}
