package main

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strconv"
	"yatc/internal"
	"yatc/user/internal"
	"yatc/user/internal/followers"
	iusers "yatc/user/internal/users"
	"yatc/user/pkg/users"
)

func main() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	config := internal.NewConfig("user/config/config.yaml", logger)

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

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		logger.Fatal("port not a int", zap.String("port", config.Port))
	}

	server := internal.NewServer(logger, port)
	server.Router.Route("/", userApi.ConfigureRouter)

	server.StartAndWait()
}
