package main

import (
	dapr "github.com/dapr/go-sdk/client"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"yatc/status/internal"
)

func main() {
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	publisher := statuses.NewDaprStatusPublisher(client)
	repo := statuses.NewInMemoryRepo()
	service := statuses.NewStatusService(repo, publisher)
	api := statuses.NewStatusApi(service)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/", api.ConfigureRouter)

	err = http.ListenAndServe(":8082", r)
	if err != nil {
		panic("Oh no!")
	}
}
