package followers

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"yatc/internal"

	"github.com/google/uuid"
	"yatc/user/pkg/users"
)

var ctx = context.Background()

type mockRepo struct {
	users map[uuid.UUID]users.User
}

func newMockRepo() *mockRepo {
	return &mockRepo{users: make(map[uuid.UUID]users.User)}
}

func (repo *mockRepo) Get(id uuid.UUID) (users.User, error) {
	user, ok := repo.users[id]
	if !ok {
		return users.User{}, internal.NotFoundError(id)
	}
	return user, nil
}

func (repo *mockRepo) Save(user users.User) (users.User, error) {
	repo.users[user.Id] = user
	return user, nil
}

func (repo *mockRepo) List() ([]users.User, error) {
	panic("implement me")
}

func (repo *mockRepo) Delete(userId uuid.UUID) (users.User, error) {
	panic("implement me")
}

func contains(s []users.User, e string) bool {
	for _, a := range s {
		if a.Name == e {
			return true
		}
	}
	return false
}

func TestService_GetFollowers(t *testing.T) {
	mockUsers := []users.User{
		{
			Id:        uuid.New(),
			Name:      "User 1",
			Followers: internal.Ptr(internal.NewSet[uuid.UUID]()),
			Followees: internal.Ptr(internal.NewSet[uuid.UUID]()),
		},
		{
			Id:        uuid.New(),
			Name:      "User 2",
			Followers: internal.Ptr(internal.NewSet[uuid.UUID]()),
			Followees: internal.Ptr(internal.NewSet[uuid.UUID]()),
		},
		{
			Id:        uuid.New(),
			Name:      "User 3",
			Followers: internal.Ptr(internal.NewSet[uuid.UUID]()),
			Followees: internal.Ptr(internal.NewSet[uuid.UUID]()),
		},
	}
	mockUsers[0].Followers.Add(mockUsers[1].Id)
	mockUsers[0].Followers.Add(mockUsers[2].Id)

	mockRepo := newMockRepo()
	for _, user := range mockUsers {
		_, err := mockRepo.Save(user)
		if err != nil {
			t.Error(err)
		}
	}

	service := NewFollowerService(mockRepo)

	followers, err := service.GetFollowers(ctx, mockUsers[0].Id)
	assert.NoError(t, err)

	assert.Len(t, followers, 2, "expected 2 followers, got %d", len(followers))
	assert.True(t, contains(followers, "User 2"), "expected followers to contain User 2. %v", followers)
	assert.True(t, contains(followers, "User 2"), "expected followers to contain User 3. %v", followers)
}

func TestService_GetFollowees(t *testing.T) {
	mockUsers := []users.User{
		{
			Id:        uuid.New(),
			Name:      "User 1",
			Followers: internal.Ptr(internal.NewSet[uuid.UUID]()),
			Followees: internal.Ptr(internal.NewSet[uuid.UUID]()),
		},
		{
			Id:        uuid.New(),
			Name:      "User 2",
			Followers: internal.Ptr(internal.NewSet[uuid.UUID]()),
			Followees: internal.Ptr(internal.NewSet[uuid.UUID]()),
		},
		{
			Id:        uuid.New(),
			Name:      "User 3",
			Followers: internal.Ptr(internal.NewSet[uuid.UUID]()),
			Followees: internal.Ptr(internal.NewSet[uuid.UUID]()),
		},
	}
	mockUsers[0].Followees.Add(mockUsers[1].Id)
	mockUsers[0].Followees.Add(mockUsers[2].Id)

	mockRepo := newMockRepo()
	for _, user := range mockUsers {
		_, err := mockRepo.Save(user)
		if err != nil {
			t.Error(err)
		}
	}

	service := NewFollowerService(mockRepo)

	followers, err := service.GetFollowees(ctx, mockUsers[0].Id)
	assert.NoError(t, err)

	assert.Len(t, followers, 2, "expected 2 followees, got %d", len(followers))
	assert.True(t, contains(followers, "User 2"), "expected followees to contain User 2. %v", followers)
	assert.True(t, contains(followers, "User 2"), "expected followees to contain User 3. %v", followers)
}
