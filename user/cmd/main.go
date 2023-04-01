package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ilyakaznacheev/cleanenv"
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
	defer func(logger *zap.Logger) {
		_ = logger.Sync()
	}(logger)

	var config internal.Config
	err := cleanenv.ReadConfig("user/config/config.yaml", &config)
	if err != nil {
		description, _ := cleanenv.GetDescription(&config, nil)
		logger.Info("Config usage" + description)
		logger.Warn("couldn't read config, using env as fallback", zap.Error(err))
		err := cleanenv.ReadEnv(&config)
		if err != nil {
			logger.Fatal("couldn't init config with config.yaml or env", zap.Error(err))
		}
	}

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
	r.Use(internal.ZapLogger(logger))
	r.Route("/", userApi.ConfigureRouter)

	port, err := strconv.Atoi(config.Port)
	if err != nil {
		logger.Fatal("port not a int", zap.String("port", config.Port))
	}
	server := internal.NewServer(logger, port, r)
	server.StartAndWait()
}
