package users

import (
	"github.com/google/uuid"
	"yatc/internal"
	"yatc/user/pkg/users"
)

type Service struct {
	repo Repository
}

func NewUserService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (userService *Service) GetUsers() ([]users.User, error) {
	return userService.repo.List()
}

func (userService *Service) GetUser(uuid uuid.UUID) (users.User, error) {
	return userService.repo.Get(uuid)
}

func (userService *Service) CreateUser(user users.User) (users.User, error) {
	user.Followers = internal.Ptr(internal.NewSet[uuid.UUID]())
	user.Followees = internal.Ptr(internal.NewSet[uuid.UUID]())
	return userService.repo.Save(user)
}

func (userService *Service) DeleteUser(uuid uuid.UUID) (users.User, error) {
	return userService.repo.Delete(uuid)
}
