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

func UserResponseToUser(userResponse api.UserResponse) users.User {
	return users.User{
		Id:   userResponse.Id,
		Name: userResponse.Username,
	}
}

func NewFollowerClient(config internal.DaprConfig) *FollowerClient {
	//TODO: Could use NewClientWithResponses
	server := fmt.Sprintf("%s:%s", config.Host, config.HttpPort)

	traceRequestFn := api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		value := ctx.Value(internal.ContextKeyTraceParent)
		if value != nil {
			trace, ok := value.(string)
			if ok {
				req.Header.Add("Traceparent", trace)
			}
		}
		return nil
	})

	daprHeaderFn := api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Add("dapr-app-id", config.AppIds.User)
		return nil
	})

	httpClient, _ := api.NewClient(server, traceRequestFn, daprHeaderFn)

	return &FollowerClient{httpClient}
}

func ToClientError(response *http.Response, err error) *internal.ClientError {
	if err != nil {
		return internal.NewClientError(nil, err)
	}

	if response.StatusCode != http.StatusOK {
		var errorResponse internal.ErrorResponse
		err := render.DecodeJSON(response.Body, &errorResponse)
		if err != nil {
			return internal.NewClientError(nil, err)
		}

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

func (client *FollowerClient) GetFollowers(ctx context.Context, userId uuid.UUID) ([]users.User, error) {
	response, err := client.httpClient.GetFollowers(ctx, userId)
	clientError := ToClientError(response, err)
	if clientError != nil {
		return nil, clientError
	}

	var usersResponse api.UsersResponse
	err = render.DecodeJSON(response.Body, &usersResponse)
	if err != nil {
		return nil, err
	}

	allUsers := make([]users.User, 0)
	for _, user := range usersResponse.Users {
		user := UserResponseToUser(user)
		allUsers = append(allUsers, user)
	}

	return allUsers, nil
}

func (client *FollowerClient) GetFollowees(ctx context.Context, userId uuid.UUID) ([]users.User, error) {
	response, err := client.httpClient.GetFollowees(ctx, userId)
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

func (client *FollowerClient) FollowUser(ctx context.Context, userToFollowId uuid.UUID, userWhichFollowsId uuid.UUID) (users.User, error) {
	//TODO: Should somehow use JWT Here instead of header maybe, do request through api gateway? and api gateway pass-through Bearer token
	body := api.FollowUserJSONRequestBody{Id: userWhichFollowsId}
	params := api.FollowUserParams{XUser: userToFollowId}
	response, err := client.httpClient.FollowUser(ctx, userToFollowId, &params, body)
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

func (client *FollowerClient) UnfollowUser(ctx context.Context, userToFollowId uuid.UUID, userWhichFollowsId uuid.UUID) error {
	//TODO: Should somehow use JWT Here instead of header maybe, do request through api gateway? and api gateway pass-through Bearer token
	params := api.UnfollowUserParams{XUser: userToFollowId}
	response, err := client.httpClient.UnfollowUser(ctx, userToFollowId, userWhichFollowsId, &params)
	clientError := ToClientError(response, err)
	if clientError != nil {
		return clientError
	}
	return nil
}
