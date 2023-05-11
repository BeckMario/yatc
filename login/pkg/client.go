package login

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
	"yatc/internal"
)

type LoginClient struct {
	url string
}

func NewLoginClient(config internal.DaprConfig) *LoginClient {
	server := fmt.Sprintf("%s:%s", config.Host, config.HttpPort)
	return &LoginClient{
		url: server,
	}
}

type Response struct {
	AccessToken string `json:"access_token"`
}

type Request struct {
	Username string `json:"username"`
	Id       string `json:"id"`
}

func (client *LoginClient) Login(username string, userId uuid.UUID) (string, error) {

	request := Request{
		Username: username,
		Id:       userId.String(),
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/login", client.url), strings.NewReader(string(payload)))

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("dapr-app-id", "login-service")

	res, err := http.DefaultClient.Do(req)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)
	clientError := internal.ToClientError(res, err)
	if clientError != nil {
		return "", clientError
	}

	var response Response
	err = render.DecodeJSON(res.Body, &response)
	if err != nil {
		return "", err
	}
	return response.AccessToken, nil
}
