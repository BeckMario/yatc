package main

import (
	dapr "github.com/dapr/go-sdk/client"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"strconv"
	"yatc/internal"
	statuses "yatc/status/internal"
)

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	config := internal.NewConfig("status/config/config.yaml", logger)

	client, err := dapr.NewClientWithPort(config.Dapr.GrpcPort)
	if err != nil {
		logger.Fatal("cant connect to dapr sidecar", zap.Error(err))
	}
	defer client.Close()

	/*	db, err := sqlx.Connect("postgres", config.Database)
		if err != nil {
			logger.Fatal("cant connect to database", zap.Error(err))
		}
		defer db.Close()

		//TODO: Temporary use migration?
		schema := `CREATE TABLE IF NOT EXISTS statuses (
				id UUID PRIMARY KEY,
				content TEXT,
				user_id UUID
			);`

		_, err = db.Exec(schema)
		if err != nil {
			logger.Fatal("cant apply default scheme to database", zap.Error(err))
		}*/

	publisher := statuses.NewDaprStatusPublisher(client, config.Dapr.PubSub)
	//repo := statuses.NewInMemoryRepo()
	repo := statuses.NewDaprStateStore(client, config.Dapr.StateStore) //statuses.NewPostgresRepo(db)
	service := statuses.NewStatusService(repo, publisher)
	api := statuses.NewStatusApi(service)

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		logger.Fatal("port not a int", zap.String("port", config.Port))
	}

	server := internal.NewServer(logger, port)
	server.Router.Route("/", api.ConfigureRouter)

	server.StartAndWait()
}
