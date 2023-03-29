package main

import (
	dapr "github.com/dapr/go-sdk/client"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"yatc/internal"
	"yatc/status/internal"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	client, err := dapr.NewClient()
	if err != nil {
		logger.Fatal("cant connect to dapr sidecar", zap.Error(err))
	}
	defer client.Close()

	//TODO: Use Config
	db, err := sqlx.Connect("postgres", "postgres://postgres:password@db:5432/postgres?sslmode=disable&connect_timeout=5") //"user=postgres dbname=postgres password=password sslmode=disable")
	if err != nil {
		logger.Fatal("cant connect to database", zap.Error(err))
	}

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

	publisher := statuses.NewDaprStatusPublisher(client)
	//repo := statuses.NewInMemoryRepo()
	repo := statuses.NewPostgresRepo(db)
	service := statuses.NewStatusService(repo, publisher)
	api := statuses.NewStatusApi(service)

	r := chi.NewRouter()
	r.Use(internal.ZapLogger(logger))
	r.Route("/", api.ConfigureRouter)

	server := internal.NewServer(logger, 8082, r)
	server.StartAndWait()
}
