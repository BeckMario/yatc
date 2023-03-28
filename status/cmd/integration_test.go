package main

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	statuses "yatc/status/pkg"
)

func Test_GetStatus(t *testing.T) {
	client, _ := statuses.NewClientWithResponses("http://status-dapr:3500", statuses.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Add("dapr-app-id", "status-service")
		return nil
	}))

	createdStatus := statuses.CreateStatusRequest{
		Content: "Test",
		UserId:  uuid.New(),
	}

	response, err := client.CreateStatusWithResponse(context.Background(), createdStatus)
	if err != nil {
		panic(err)
	}

	status, err := client.GetStatusWithResponse(context.Background(), (*response.JSON201).Id)
	if err != nil {
		return
	}

	statusResponse := *status.JSON200

	assert.Equal(t, createdStatus.UserId, statusResponse.UserId)
	assert.Equal(t, createdStatus.Content, statusResponse.Content)
	assert.Equal(t, response.JSON201.Id, statusResponse.Id)
}
