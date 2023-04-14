package main

import (
	dapr "github.com/dapr/go-sdk/client"
	"go.uber.org/zap"
	"mime"
	"strconv"
	"yatc/internal"
	media "yatc/media/internal"
)

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	config := internal.NewConfig("media/config/config.yaml", logger)

	client, err := dapr.NewClientWithPort(config.Dapr.GrpcPort)
	if err != nil {
		logger.Fatal("cant connect to dapr sidecar", zap.Error(err))
	}
	defer client.Close()

	// Neded because scratch image does not have mime types table
	_ = mime.AddExtensionType(".mp4", "video/mp4")

	s3 := media.NewDaprS3(client, config.Dapr.S3)
	publisher := media.NewDaprMediaPublisher(client, config.Dapr.PubSub)
	service := media.NewMediaService(s3, publisher)
	api := media.NewMediaApi(service)

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		logger.Fatal("port not a int", zap.String("port", config.Port))
	}

	server := internal.NewServer(logger, port)
	server.Router.Route("/", api.ConfigureRouter)

	server.StartAndWait()
}
