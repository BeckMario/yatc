package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/ilyakaznacheev/cleanenv"
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

	var config internal.Config
	err := cleanenv.ReadConfig("timeline/config/config.yaml", &config)
	if err != nil {
		description, _ := cleanenv.GetDescription(&config, nil)
		logger.Info("Config usage" + description)
		logger.Warn("couldn't read config, using env as fallback", zap.Error(err))
		err := cleanenv.ReadEnv(&config)
		if err != nil {
			logger.Fatal("couldn't init config with config.yaml or env", zap.Error(err))
		}
	}

	repo := timelines.NewInMemoryRepo()
	client := followers.NewFollowerClient(config.Dapr)
	service := timelines.NewTimelineService(repo, client)
	api := timelines.NewTimelineApi(service)

	r := chi.NewRouter()
	r.Use(internal.ZapLogger(logger))
	r.Route("/", api.ConfigureRouter)

	subscriber := statuses.NewDaprStatusSubscriber(r, logger, config.Dapr.PubSub)
	subscriber.Subscribe(func(status statuses.Status) {
		err := service.UpdateTimelines(status.UserId, status)
		if err != nil {
			logger.Error("updateing timelines", zap.Error(err), zap.Any("status", status))
		}
	})

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		logger.Fatal("port not a int", zap.String("port", config.Port))
	}
	server := internal.NewServer(logger, port, r)
	server.StartAndWait()
}
