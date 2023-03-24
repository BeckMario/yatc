package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"yatc/user/internal"
	"yatc/user/internal/followers"
	iusers "yatc/user/internal/users"
	"yatc/user/pkg/users"
)

func main() {
	userRepo := iusers.NewInMemoryRepo()
	_, _ = userRepo.Save(users.User{
		Id:        uuid.MustParse("dc52828f-9c08-4e38-ace0-bf2bd87bfff6"),
		Name:      "Hans",
		Followers: map[uuid.UUID]struct{}{},
		Followees: map[uuid.UUID]struct{}{},
	})

	_, _ = userRepo.Save(users.User{
		Id:        uuid.MustParse("e0758810-9119-4b8e-b3b8-53c5959d0bee"),
		Name:      "Peter",
		Followers: map[uuid.UUID]struct{}{},
		Followees: map[uuid.UUID]struct{}{},
	})

	userService := iusers.NewUserService(userRepo)
	followerService := followers.NewFollowerService(userRepo)
	userApi := api.NewUserApi(userService, followerService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/", userApi.ConfigureRouter)

	server := &http.Server{
		Addr:    ":8080",
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
