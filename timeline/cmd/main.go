package main

import (
	"context"
	dapr "github.com/dapr/go-sdk/client"
	"go.uber.org/zap"
	"strconv"
	"yatc/internal"
	"yatc/status/pkg"
	"yatc/timeline/internal"
	"yatc/user/pkg/followers"
)

func main() {
	logger, sync := internal.NewZapLogger()
	defer sync(logger)

	config := internal.NewConfig("timeline/config/config.yaml", logger)

	client, err := dapr.NewClientWithPort(config.Dapr.GrpcPort)
	if err != nil {
		logger.Fatal("cant connect to dapr sidecar", zap.Error(err))
	}
	defer client.Close()

	repo := timelines.NewDaprRepo(client, config.Dapr.StateStore) //timelines.NewInMemoryRepo()
	followerClient := followers.NewFollowerClient(config.Dapr)
	service := timelines.NewTimelineService(repo, followerClient, client, config.Dapr.PubSub)
	api := timelines.NewTimelineApi(service)

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		logger.Fatal("port not a int", zap.String("port", config.Port))
	}
	server := internal.NewServer(logger, port)

	server.Router.Route("/", api.ConfigureRouter)

	subscriber := statuses.NewDaprStatusSubscriber(server.Router, logger, config.Dapr.PubSub)
	subscriber.Subscribe(func(ctx context.Context, status statuses.Status) {
		err := service.UpdateTimelines(ctx, status.UserId, status)
		if err != nil {
			logger.Error("updating timelines", zap.Error(err), zap.Any("status", status))
		}
	})

	server.StartAndWait()
}
