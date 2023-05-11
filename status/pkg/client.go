package statuses

import (
	"context"
	"fmt"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"net/http"
	"yatc/internal"
)

type StatusClient struct {
	httpClient ClientInterface
}

func NewStatusClient(config internal.DaprConfig) *StatusClient {
	server := fmt.Sprintf("%s:%s", config.Host, config.HttpPort)

	traceRequestFn := WithRequestEditorFn(internal.OapiClientTraceRequestFn())
	authRequestFn := WithRequestEditorFn(internal.OapiClientAuthRequestFn())

	daprHeaderFn := WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Add("dapr-app-id", config.AppIds.User)
		return nil
	})

	httpClient, _ := NewClient(server, traceRequestFn, daprHeaderFn, authRequestFn)

	return &StatusClient{httpClient}
}

func (client *StatusClient) GetStatuses(userId uuid.UUID) ([]Status, error) {
	panic("implement me")
}

func (client *StatusClient) GetStatus(statusId uuid.UUID) (Status, error) {
	panic("implement me")
}

func (client *StatusClient) CreateStatus(ctx context.Context, status Status) (Status, error) {
	body := CreateStatusRequest{
		Content: status.Content,
	}

	response, err := client.httpClient.CreateStatus(ctx, &CreateStatusParams{XUser: status.UserId}, body)
	clientError := internal.ToClientError(response, err)
	if clientError != nil {
		return Status{}, clientError
	}

	var statusResponse StatusResponse
	err = render.DecodeJSON(response.Body, &statusResponse)
	if err != nil {
		return Status{}, err
	}

	return Status{Id: statusResponse.Id, Content: statusResponse.Content}, nil
}

func (client *StatusClient) DeleteStatus(statusId uuid.UUID) (Status, error) {
	panic("implement me")
}
