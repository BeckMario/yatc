package main

import (
	dapr "github.com/dapr/go-sdk/client"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strconv"
	"yatc/internal"
	"yatc/user/internal"
	"yatc/user/internal/followers"
	iusers "yatc/user/internal/users"
	"yatc/user/pkg/users"
)

func main() {
	logger, sync := internal.NewZapLogger()
	defer sync(logger)

	config := internal.NewConfig("user/config/config.yaml", logger)

	client, err := dapr.NewClientWithPort(config.Dapr.GrpcPort)
	if err != nil {
		logger.Fatal("cant connect to dapr sidecar", zap.Error(err))
	}
	defer client.Close()

	userRepo := iusers.NewDaprRepo(client, config.Dapr.StateStore)
	_, err = userRepo.Save(users.User{
		Id:        uuid.MustParse("dc52828f-9c08-4e38-ace0-bf2bd87bfff6"),
		Name:      "Hans",
		Followers: internal.Ptr(internal.NewSet[uuid.UUID]()),
		Followees: internal.Ptr(internal.NewSet[uuid.UUID]()),
	})

	if err != nil {
		logger.Error("error", zap.Error(err))
	}

	_, err = userRepo.Save(users.User{
		Id:        uuid.MustParse("e0758810-9119-4b8e-b3b8-53c5959d0bee"),
		Name:      "Peter",
		Followers: internal.Ptr(internal.NewSet[uuid.UUID]()),
		Followees: internal.Ptr(internal.NewSet[uuid.UUID]()),
	})

	if err != nil {
		logger.Fatal("error", zap.Error(err))
	}

	userService := iusers.NewUserService(userRepo)
	followerService := followers.NewFollowerService(userRepo)
	userApi := api.NewUserApi(userService, followerService)

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		logger.Fatal("port not a int", zap.String("port", config.Port))
	}

	server := internal.NewServer(logger, port)
	server.Router.Route("/", userApi.ConfigureRouter)

	server.StartAndWait()
}
