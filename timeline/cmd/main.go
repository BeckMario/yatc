package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"yatc/internal"
	"yatc/status/pkg"
	"yatc/timeline/internal"
	"yatc/user/pkg/followers"
)

func main() {
	repo := timelines.NewInMemoryRepo()
	client := followers.NewFollowerClient("http://localhost:3501")
	service := timelines.NewTimelineService(repo, client)
	api := timelines.NewTimelineApi(service)

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/", api.ConfigureRouter)

	subscriber := statuses.NewDaprTweetSubscriber(r, logger)
	subscriber.Subscribe(func(status statuses.Status) {
		err := service.UpdateTimelines(status.UserId, status)
		if err != nil {
			logger.Error("updateing timelines", zap.Error(err), zap.Any("status", status))
		}
	})

	server := internal.NewServer(logger, 8081, r)
	server.StartAndWait()
}
