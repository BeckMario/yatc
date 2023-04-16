package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
	statuses "yatc/status/pkg"
)

func Test_GetStatus(t *testing.T) {
	serverAddr := os.Getenv("STATUS_SERVICE_ADDR")
	if serverAddr == "" {
		t.Skipf("set STATUS_SERVICE_ADDR to run this integration test")
	}

	client, _ := statuses.NewClientWithResponses(serverAddr, statuses.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Add("dapr-app-id", "status-service")
		return nil
	}))

	createdStatus := statuses.CreateStatusRequest{
		Content: "Test",
	}
	userId := uuid.New()
	params := statuses.CreateStatusParams{XUser: userId}

	response, err := client.CreateStatusWithResponse(context.Background(), &params, createdStatus)
	if err != nil {
		panic(err)
	}

	status, err := client.GetStatusWithResponse(context.Background(), (*response.JSON201).Id)
	if err != nil {
		return
	}

	statusResponse := *status.JSON200

	assert.Equal(t, userId, statusResponse.UserId)
	assert.Equal(t, createdStatus.Content, statusResponse.Content)
	assert.Equal(t, response.JSON201.Id, statusResponse.Id)
}
