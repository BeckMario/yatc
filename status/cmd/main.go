package main

import (
	"context"
	"errors"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	server := &http.Server{
		Addr:    ":8082",
		Handler: r,
	}

	go func() {
		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")
}
