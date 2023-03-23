package users

import "github.com/google/uuid"

type User struct {
	Id        uuid.UUID
	Name      string
	Followers map[uuid.UUID]struct{}
	Followees map[uuid.UUID]struct{}
}

type UserService interface {
	GetUsers() ([]User, error)
	GetUser(userId uuid.UUID) (User, error)
	CreateUser(user User) (User, error)
	DeleteUser(userId uuid.UUID) (User, error)
}
