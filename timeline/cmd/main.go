package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"yatc/status/pkg"
	"yatc/timeline/internal"
	"yatc/user/pkg/followers"
)

func main() {
	repo := timelines.NewInMemoryRepo()
	client := followers.NewFollowerClient("http://localhost:3501")
	service := timelines.NewTweetService(repo, client)
	api := timelines.NewTimelineApi(service)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/", api.ConfigureRouter)

	subscriber := statuses.NewDaprTweetSubscriber(r)
	subscriber.Subscribe(func(status statuses.Status) {
		fmt.Println("Got Tweet:", status)
		timeline, err := service.UpdateTimeline(status.UserId, status)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			return
		}
		fmt.Println(timeline)
	})

	server := &http.Server{
		Addr:    ":8081",
		Handler: r,
	}

	go func() {
		err := server.ListenAndServe()
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
