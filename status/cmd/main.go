package main

import (
	dapr "github.com/dapr/go-sdk/client"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	istatuses "yatc/status/internal"
	"yatc/status/pkg"
)

func main() {
	client, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	var statusPublisher istatuses.Publisher = istatuses.NewDaprStatusPublisher(client)
	var statusRepo istatuses.Repository = istatuses.NewInMemoryRepo()
	var statusService statuses.Service = istatuses.NewStatusService(statusRepo, statusPublisher)
	var statusApi = istatuses.NewStatusApi(statusService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/", statusApi.ConfigureRouter)

	err = http.ListenAndServe(":8082", r)
	if err != nil {
		panic("Oh no!")
	}
}
