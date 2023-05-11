package timelines

import (
	"context"
	"fmt"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"net/http"
	"yatc/internal"
	statuses "yatc/status/pkg"
)

type TimelineClient struct {
	httpClient ClientInterface
}

func NewTimelineClient(config internal.DaprConfig) *TimelineClient {
	server := fmt.Sprintf("%s:%s", config.Host, config.HttpPort)

	traceRequestFn := WithRequestEditorFn(internal.OapiClientTraceRequestFn())
	authRequestFn := WithRequestEditorFn(internal.OapiClientAuthRequestFn())

	daprHeaderFn := WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Add("dapr-app-id", config.AppIds.User)
		return nil
	})

	httpClient, _ := NewClient(server, traceRequestFn, daprHeaderFn, authRequestFn)

	return &TimelineClient{httpClient}
}

func StatusResponsesToStatuses(responses []statuses.StatusResponse) []statuses.Status {
	allStatuses := make([]statuses.Status, 0)
	for _, statusResponse := range responses {
		allStatuses = append(allStatuses, statuses.Status{
			Id:      statusResponse.Id,
			Content: statusResponse.Content,
			UserId:  statusResponse.UserId,
		})
	}
	return allStatuses
}

func (client *TimelineClient) GetTimeline(ctx context.Context, userId uuid.UUID) (Timeline, error) {
	response, err := client.httpClient.GetTimeline(ctx, userId)
	clientError := internal.ToClientError(response, err)
	if clientError != nil {
		return Timeline{}, clientError
	}

	var timelineResponse TimelineResponse
	err = render.DecodeJSON(response.Body, &timelineResponse)
	if err != nil {
		return Timeline{}, err
	}

	return Timeline{
		UserId:   timelineResponse.Id,
		Statuses: StatusResponsesToStatuses(timelineResponse.Statuses),
	}, nil
}

func (client *TimelineClient) UpdateTimelines(ctx context.Context, userId uuid.UUID, status statuses.Status) error {
	panic("implement me")
}
