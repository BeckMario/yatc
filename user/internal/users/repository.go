package users

import (
	"github.com/google/uuid"
	"yatc/internal"
	"yatc/user/pkg/users"
)

type Repository interface {
	List() ([]users.User, error)
	Get(userId uuid.UUID) (users.User, error)
	Delete(userId uuid.UUID) (users.User, error)
	Save(status users.User) (users.User, error)
}

type InMemoryRepo struct {
	Users map[uuid.UUID]users.User
}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{Users: map[uuid.UUID]users.User{}}
}

func (repo InMemoryRepo) List() ([]users.User, error) {
	values := make([]users.User, 0, len(repo.Users))

	for _, value := range repo.Users {
		values = append(values, value)
	}

	return values, nil
}

func (repo InMemoryRepo) Get(userId uuid.UUID) (users.User, error) {
	user, ok := repo.Users[userId]
	if !ok {
		return users.User{}, internal.NotFoundError(userId)
	}
	return user, nil
}

func (repo InMemoryRepo) Delete(userId uuid.UUID) (users.User, error) {
	user, exists := repo.Users[userId]
	if !exists {
		return users.User{}, internal.NotFoundError(userId)
	}
	delete(repo.Users, userId)
	return user, nil
}

func (repo InMemoryRepo) Save(user users.User) (users.User, error) {
	repo.Users[user.Id] = user
	return user, nil
}
