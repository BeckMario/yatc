package statuses

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	statuses "yatc/status/pkg"
)

type MockRepository struct {
	CreateCalled bool
}

func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

func (repo *MockRepository) Create(status statuses.Status) (statuses.Status, error) {
	repo.CreateCalled = true
	return status, nil
}

func (repo *MockRepository) Get(statusId uuid.UUID) (statuses.Status, error) {
	return statuses.Status{}, nil
}

func (repo *MockRepository) List() ([]statuses.Status, error) {
	return []statuses.Status{}, nil
}

func (repo *MockRepository) Delete(statusId uuid.UUID) (statuses.Status, error) {
	return statuses.Status{}, nil
}

type MockPublisher struct {
	PublishCalled bool
}

func NewMockPublisher() *MockPublisher {
	return &MockPublisher{}
}

func (publisher *MockPublisher) Publish(status statuses.Status) error {
	publisher.PublishCalled = true
	return nil
}

func TestService_CreateStatus(t *testing.T) {
	// GIVEN
	repo := NewMockRepository()
	publisher := NewMockPublisher()
	service := NewStatusService(repo, publisher)
	status := statuses.Status{Id: uuid.New(), Content: "test status", UserId: uuid.New()}

	// WHEN
	createdStatus, err := service.CreateStatus(context.Background(), status)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, status, createdStatus)
	assert.True(t, repo.CreateCalled)
	assert.True(t, publisher.PublishCalled)
}
