package statuses

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	statuses "yatc/status/pkg"
)

func TestInMemoryRepo_Create(t *testing.T) {
	// GIVEN
	repo := NewInMemoryRepo()
	status := statuses.Status{Id: uuid.New(), Content: "test status", UserId: uuid.New()}

	// WHEN
	createdStatus, err := repo.Create(status)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, status, createdStatus)
}

func TestInMemoryRepo_Get(t *testing.T) {
	// GIVEN
	repo := NewInMemoryRepo()
	status := statuses.Status{Id: uuid.New(), Content: "test status", UserId: uuid.New()}
	_, err := repo.Create(status)
	assert.NoError(t, err)

	// WHEN
	fetchedStatus, err := repo.Get(status.Id)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, status, fetchedStatus)
}

func TestInMemoryRepo_List(t *testing.T) {
	// GIVEN
	repo := NewInMemoryRepo()
	status := statuses.Status{Id: uuid.New(), Content: "test status", UserId: uuid.New()}
	_, err := repo.Create(status)
	assert.NoError(t, err)

	// WHEN
	statusList, err := repo.List()

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, 1, len(statusList))
	assert.Equal(t, status, statusList[0])
}

func TestInMemoryRepo_Delete(t *testing.T) {
	// GIVEN
	repo := NewInMemoryRepo()
	status := statuses.Status{Id: uuid.New(), Content: "test status", UserId: uuid.New()}
	_, err := repo.Create(status)
	assert.NoError(t, err)

	// WHEN
	deletedStatus, err := repo.Delete(status.Id)

	// THEN
	assert.Nil(t, err)
	assert.Equal(t, status, deletedStatus)

	statusList, err := repo.List()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(statusList))

	fetchedStatus, err := repo.Get(status.Id)
	assert.Error(t, err)
	assert.Empty(t, fetchedStatus)
}

func TestInMemoryRepo_Delete_NonExistentStatus(t *testing.T) {
	// GIVEN
	repo := NewInMemoryRepo()
	nonExistentStatusID := uuid.New()

	// WHEN
	deletedStatus, err := repo.Delete(nonExistentStatusID)

	// THEN
	assert.Error(t, err)
	assert.Empty(t, deletedStatus)
}

func TestInMemoryRepo_Get_NonExistentStatus(t *testing.T) {
	// GIVEN
	repo := NewInMemoryRepo()
	nonExistentStatusID := uuid.New()

	// WHEN
	fetchedStatus, err := repo.Get(nonExistentStatusID)

	// THEN
	assert.Error(t, err)
	assert.Empty(t, fetchedStatus)
}

func TestInMemoryRepo_Create_DuplicateStatus(t *testing.T) {
	// GIVEN
	repo := NewInMemoryRepo()
	status := statuses.Status{Id: uuid.New(), Content: "test status", UserId: uuid.New()}
	_, err := repo.Create(status)
	assert.NoError(t, err)

	// WHEN
	createdStatus, err := repo.Create(status)

	// THEN
	assert.Error(t, err)
	assert.Empty(t, createdStatus)
}
