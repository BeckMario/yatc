package main

import (
	dapr "github.com/dapr/go-sdk/client"
	"github.com/go-chi/chi/v5"
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
	"strconv"
	"yatc/internal"
	media "yatc/media/internal"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	var config internal.Config
	err := cleanenv.ReadConfig("media/config/config.yaml", &config)

	if err != nil {
		description, _ := cleanenv.GetDescription(&config, nil)
		logger.Info("Config usage" + description)
		logger.Warn("couldn't read config, using env as fallback", zap.Error(err))
		err := cleanenv.ReadEnv(&config)
		if err != nil {
			logger.Fatal("couldn't init config with config.yaml or env", zap.Error(err))
		}
	}

	client, err := dapr.NewClientWithPort(config.Dapr.GrpcPort)
	if err != nil {
		logger.Fatal("cant connect to dapr sidecar", zap.Error(err))
	}
	defer client.Close()

	service := media.NewMediaService(client)
	api := media.NewMediaApi(service)

	r := chi.NewRouter()
	r.Use(internal.ZapLogger(logger))
	r.Route("/", api.ConfigureRouter)

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		logger.Fatal("port not a int", zap.String("port", config.Port))
	}
	server := internal.NewServer(logger, port, r)
	server.StartAndWait()
}
