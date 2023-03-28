package followers

import (
	"context"
	"fmt"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"net/http"
	"yatc/internal"
	api "yatc/user/pkg"
	"yatc/user/pkg/users"
)

type FollowerClient struct {
	httpClient api.ClientInterface
}

func NewFollowerClient(server string) *FollowerClient {
	//TODO: Could use NewClientWithResponses
	httpClient, _ := api.NewClient(server, api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		//TODO: Config?
		req.Header.Add("dapr-app-id", "user-service")
		return nil
	}))

	return &FollowerClient{httpClient}
}

func ToClientError(response *http.Response, err error) *internal.ClientError {
	if err != nil {
		fmt.Printf("Received Client Error: %s\n", err)
		return internal.NewClientError(nil, err)
	}
	if response.StatusCode != http.StatusOK {
		fmt.Printf("StatusCode not ok: %d\n", response.StatusCode)
		var errorResponse internal.ErrorResponse
		err := render.DecodeJSON(response.Body, &errorResponse)
		if err != nil {
			return internal.NewClientError(nil, err)
		}

		fmt.Printf("ErrorResponse: %v\n", errorResponse)

		if response.StatusCode == http.StatusNotFound {
			id, err := uuid.Parse(errorResponse.Message)
			if err != nil {
				return internal.NewClientError(&errorResponse, err)
			}
			return internal.NewClientError(&errorResponse, internal.NotFoundError(id))
		}

		return internal.NewClientError(&errorResponse, nil)
	}
	return nil
}

func (client *FollowerClient) GetFollowers(userId uuid.UUID) ([]users.User, error) {
	response, err := client.httpClient.GetFollowers(context.Background(), userId)
	clientError := ToClientError(response, err)
	if clientError != nil {
		return nil, clientError
	}
	var allUsers []users.User
	err = render.DecodeJSON(response.Body, &allUsers)
	if err != nil {
		return nil, err
	}
	return allUsers, nil
}

func (client *FollowerClient) GetFollowees(userId uuid.UUID) ([]users.User, error) {
	response, err := client.httpClient.GetFollowees(context.Background(), userId)
	clientError := ToClientError(response, err)
	if clientError != nil {
		return nil, clientError
	}

	var allUsers []users.User
	err = render.DecodeJSON(response.Body, &allUsers)
	if err != nil {
		return nil, err
	}
	return allUsers, nil
}

func (client *FollowerClient) FollowUser(userToFollowId uuid.UUID, userWhichFollowsId uuid.UUID) (users.User, error) {
	body := api.FollowUserJSONRequestBody{Id: userWhichFollowsId}
	response, err := client.httpClient.FollowUser(context.Background(), userToFollowId, body)
	clientError := ToClientError(response, err)
	if clientError != nil {
		return users.User{}, clientError
	}
	var user users.User
	err = render.DecodeJSON(response.Body, user)
	if err != nil {
		return users.User{}, err
	}
	return user, nil
}

func (client *FollowerClient) UnfollowUser(userToFollowId uuid.UUID, userWhichFollowsId uuid.UUID) error {
	response, err := client.httpClient.UnfollowUser(context.Background(), userToFollowId, userWhichFollowsId)
	clientError := ToClientError(response, err)
	if clientError != nil {
		return clientError
	}
	return nil
}
