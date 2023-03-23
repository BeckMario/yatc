package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"net/http"
	api "yatc/user/internal"
	"yatc/user/internal/followers"
	iusers "yatc/user/internal/users"
	"yatc/user/pkg/users"
)

//go:generate oapi-codegen --config ../openapi/oapi-codegen-config.server.yaml ../openapi/openapi.yaml
//go:generate oapi-codegen --config ../openapi/oapi-codegen-config.client.yaml ../openapi/openapi.yaml
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
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic("Oh no!")
	}
}
