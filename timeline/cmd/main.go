package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
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

	err := http.ListenAndServe(":8081", r)
	if err != nil {
		panic("Oh no!")
	}
}
