package statuses

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	statuses "yatc/status/pkg"
)

type MockService struct {
	statuses []statuses.Status
}

func NewMockService() *MockService {
	return &MockService{}
}

func (service *MockService) GetStatuses() ([]statuses.Status, error) {
	return service.statuses, nil
}

func (service *MockService) GetStatus(statusId uuid.UUID) (statuses.Status, error) {
	for _, status := range service.statuses {
		if status.Id == statusId {
			return status, nil
		}
	}
	return statuses.Status{}, fmt.Errorf("status not found")
}

func (service *MockService) CreateStatus(status statuses.Status) (statuses.Status, error) {
	service.statuses = append(service.statuses, status)
	return status, nil
}

func (service *MockService) DeleteStatus(statusId uuid.UUID) (statuses.Status, error) {
	for i, status := range service.statuses {
		if status.Id == statusId {
			service.statuses = append(service.statuses[:i], service.statuses[i+1:]...)
			return status, nil
		}
	}
	return statuses.Status{}, fmt.Errorf("status not found")
}

func TestApi_GetStatuses(t *testing.T) {
	// GIVEN
	service := NewMockService()
	api := NewStatusApi(service)
	status1 := statuses.Status{Id: uuid.New(), Content: "test status 1", UserId: uuid.New()}
	status2 := statuses.Status{Id: uuid.New(), Content: "test status 2", UserId: uuid.New()}
	service.statuses = []statuses.Status{status1, status2}

	router := chi.NewRouter()
	api.ConfigureRouter(router)

	req, err := http.NewRequest(http.MethodGet, "/statuses", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// WHEN
	router.ServeHTTP(rr, req)

	// THEN
	assert.Equal(t, http.StatusOK, rr.Code)

	var statusResponses []statuses.StatusResponse
	err = json.NewDecoder(rr.Body).Decode(&statusResponses)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(statusResponses))
	assert.Equal(t, status1.Id, statusResponses[0].Id)
	assert.Equal(t, status1.Content, statusResponses[0].Content)
	assert.Equal(t, status1.UserId, statusResponses[0].UserId)
	assert.Equal(t, status2.Id, statusResponses[1].Id)
	assert.Equal(t, status2.Content, statusResponses[1].Content)
	assert.Equal(t, status2.UserId, statusResponses[1].UserId)
}

func TestApi_GetStatus(t *testing.T) {
	// GIVEN
	service := NewMockService()
	api := NewStatusApi(service)
	status := statuses.Status{Id: uuid.New(), Content: "test status", UserId: uuid.New()}
	service.statuses = []statuses.Status{status}

	router := chi.NewRouter()
	api.ConfigureRouter(router)

	req, err := http.NewRequest(http.MethodGet, "/statuses/"+status.Id.String(), nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// WHEN
	router.ServeHTTP(rr, req)

	// THEN
	assert.Equal(t, http.StatusOK, rr.Code)

	var statusResponse statuses.StatusResponse
	err = json.NewDecoder(rr.Body).Decode(&statusResponse)
	assert.NoError(t, err)

	assert.Equal(t, status.Id, statusResponse.Id)
	assert.Equal(t, status.Content, statusResponse.Content)
	assert.Equal(t, status.UserId, statusResponse.UserId)
}

func TestApi_CreateStatus(t *testing.T) {
	// GIVEN
	service := NewMockService()
	api := NewStatusApi(service)
	status := statuses.Status{Id: uuid.New(), Content: "test status", UserId: uuid.New()}
	createStatusRequest := statuses.CreateStatusRequest{Content: status.Content, UserId: status.UserId}
	requestBody, err := json.Marshal(createStatusRequest)
	assert.NoError(t, err)

	router := chi.NewRouter()
	api.ConfigureRouter(router)

	req, err := http.NewRequest(http.MethodPost, "/statuses", bytes.NewReader(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// WHEN
	router.ServeHTTP(rr, req)

	// THEN
	assert.Equal(t, http.StatusCreated, rr.Code)

	var statusResponse statuses.StatusResponse
	err = json.NewDecoder(rr.Body).Decode(&statusResponse)
	assert.NoError(t, err)

	assert.Equal(t, status.Content, statusResponse.Content)
	assert.Equal(t, status.UserId, statusResponse.UserId)
}

func TestApi_DeleteStatus(t *testing.T) {
	// GIVEN
	service := NewMockService()
	api := NewStatusApi(service)
	status := statuses.Status{Id: uuid.New(), Content: "test status", UserId: uuid.New()}
	service.statuses = []statuses.Status{status}

	router := chi.NewRouter()
	api.ConfigureRouter(router)

	req, err := http.NewRequest(http.MethodDelete, "/statuses/"+status.Id.String(), nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// WHEN
	router.ServeHTTP(rr, req)

	// THEN
	assert.Equal(t, http.StatusOK, rr.Code)

	var statusResponse statuses.StatusResponse
	err = json.NewDecoder(rr.Body).Decode(&statusResponse)
	assert.NoError(t, err)

	assert.Equal(t, status.Id, statusResponse.Id)
	assert.Equal(t, status.Content, statusResponse.Content)
	assert.Equal(t, status.UserId, statusResponse.UserId)
}

func TestApi_GetStatus_NonExistentStatus(t *testing.T) {
	// GIVEN
	service := NewMockService()
	api := NewStatusApi(service)
	nonExistentStatusID := uuid.New()

	router := chi.NewRouter()
	api.ConfigureRouter(router)

	req, err := http.NewRequest(http.MethodGet, "/statuses/"+nonExistentStatusID.String(), nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// WHEN
	router.ServeHTTP(rr, req)

	// THEN
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestApi_DeleteStatus_NonExistentStatus(t *testing.T) {
	// GIVEN
	service := NewMockService()
	api := NewStatusApi(service)
	nonExistentStatusID := uuid.New()

	router := chi.NewRouter()
	api.ConfigureRouter(router)

	req, err := http.NewRequest(http.MethodDelete, "/statuses/"+nonExistentStatusID.String(), nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// WHEN
	router.ServeHTTP(rr, req)

	// THEN
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestApi_CreateStatus_InvalidRequest(t *testing.T) {
	// GIVEN
	service := NewMockService()
	api := NewStatusApi(service)

	router := chi.NewRouter()
	api.ConfigureRouter(router)

	req, err := http.NewRequest(http.MethodPost, "/statuses", strings.NewReader("invalid request body"))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// WHEN
	router.ServeHTTP(rr, req)

	// THEN
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
