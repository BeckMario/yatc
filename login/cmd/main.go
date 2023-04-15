package main

import (
	"encoding/json"
	"fmt"
	dapr "github.com/dapr/go-sdk/client"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"time"
	"yatc/internal"
)

type loginRequest struct {
	Username string
	Id       string
}

type token struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
}

type user struct {
	Username string
	Id       string
}

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	config := internal.NewConfig("login/config/config.yaml", logger)

	client, err := dapr.NewClientWithPort(config.Dapr.GrpcPort)
	if err != nil {
		logger.Fatal("cant connect to dapr sidecar", zap.Error(err))
	}
	defer client.Close()

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		logger.Fatal("port not a int", zap.String("port", config.Port))
	}

	server := internal.NewServer(logger, port)
	server.Router.Post("/login", func(writer http.ResponseWriter, request *http.Request) {
		reqBytes, err := io.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		var login loginRequest
		err = json.Unmarshal(reqBytes, &login)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		url := fmt.Sprintf("%s:%s/v1.0/invoke/user-service/method/users/%s", config.Dapr.Host, config.Dapr.HttpPort, login.Id)
		response, err := http.Get(url)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if response.StatusCode != http.StatusOK {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		respBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return
		}

		var user user
		err = json.Unmarshal(respBytes, &user)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if user.Username == login.Username {
			tokens := make(map[string]token, 0)
			currentTime := time.Now()
			futureTime := currentTime.Add(time.Hour * 24)
			token := token{
				Sub: user.Id,
				Exp: futureTime.Unix(),
			}

			tokens["access_token"] = token

			jsonBytes, err := json.Marshal(tokens)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = writer.Write(jsonBytes)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	})

	server.StartAndWait()
}
