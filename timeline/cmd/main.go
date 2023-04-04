package main

import (
	"context"
	"go.uber.org/zap"
	"strconv"
	"yatc/internal"
	"yatc/status/pkg"
	"yatc/timeline/internal"
	"yatc/user/pkg/followers"
)

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	config := internal.NewConfig("timeline/config/config.yaml", logger)

	repo := timelines.NewInMemoryRepo()
	client := followers.NewFollowerClient(config.Dapr)
	service := timelines.NewTimelineService(repo, client)
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
			logger.Error("updateing timelines", zap.Error(err), zap.Any("status", status))
		}
	})

	server.StartAndWait()
}
