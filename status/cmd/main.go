package main

import (
	dapr "github.com/dapr/go-sdk/client"
	"github.com/go-chi/chi/v5"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"strconv"
	"yatc/internal"
	statuses "yatc/status/internal"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	var config internal.Config
	err := cleanenv.ReadConfig("status/config/config.yaml", &config)

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

	db, err := sqlx.Connect("postgres", config.Database)
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
	}

	publisher := statuses.NewDaprStatusPublisher(client, config.Dapr.PubSub)
	//repo := statuses.NewInMemoryRepo()
	repo := statuses.NewPostgresRepo(db)
	service := statuses.NewStatusService(repo, publisher)
	api := statuses.NewStatusApi(service)

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
