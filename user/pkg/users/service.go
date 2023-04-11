package users

import (
	"github.com/google/uuid"
	"yatc/internal"
)

type User struct {
	Id        uuid.UUID
	Name      string
	Followers *internal.Set[uuid.UUID]
	Followees *internal.Set[uuid.UUID]
}

type Service interface {
	GetUsers() ([]User, error)
	GetUser(userId uuid.UUID) (User, error)
	CreateUser(user User) (User, error)
	DeleteUser(userId uuid.UUID) (User, error)
}
