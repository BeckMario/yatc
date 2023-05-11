package users

import (
	"context"
	"fmt"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"net/http"
	"yatc/internal"
	api "yatc/user/pkg"
)

type UserClient struct {
	httpClient api.ClientInterface
}

func NewUserClient(config internal.DaprConfig) *UserClient {
	server := fmt.Sprintf("%s:%s", config.Host, config.HttpPort)

	traceRequestFn := api.WithRequestEditorFn(internal.OapiClientTraceRequestFn())
	authRequestFn := api.WithRequestEditorFn(internal.OapiClientAuthRequestFn())

	daprHeaderFn := api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Add("dapr-app-id", config.AppIds.User)
		return nil
	})

	httpClient, _ := api.NewClient(server, traceRequestFn, daprHeaderFn, authRequestFn)

	return &UserClient{httpClient}
}

func (client *UserClient) GetUsers() ([]User, error) {
	panic("implement me")
}

func (client *UserClient) GetUser(userId uuid.UUID) (User, error) {
	panic("implement me")
}

func (client *UserClient) CreateUser(user User) (User, error) {
	body := api.CreateUserJSONRequestBody{Username: user.Name}
	response, err := client.httpClient.CreateUser(context.Background(), body)
	clientError := internal.ToClientError(response, err)
	if clientError != nil {
		return User{}, clientError
	}

	var usersResponse api.UserResponse
	err = render.DecodeJSON(response.Body, &usersResponse)
	if err != nil {
		return User{}, err
	}

	return User{
		Id:   usersResponse.Id,
		Name: usersResponse.Username,
	}, nil
}

func (client *UserClient) DeleteUser(userId uuid.UUID) (User, error) {
	panic("implement me")
}
