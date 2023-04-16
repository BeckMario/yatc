package statuses

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yatc/internal"
	statuses "yatc/status/pkg"
)

type MockService struct {
	statuses []statuses.Status
}

func NewMockService() *MockService {
	return &MockService{}
}

func (service *MockService) GetStatuses(userId uuid.UUID) ([]statuses.Status, error) {
	return service.statuses, nil
}

func (service *MockService) GetStatus(statusId uuid.UUID) (statuses.Status, error) {
	for _, status := range service.statuses {
		if status.Id == statusId {
			return status, nil
		}
	}
	return statuses.Status{}, internal.NotFoundError(statusId)
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
	return statuses.Status{}, internal.NotFoundError(statusId)
}

func TestApi_GetStatuses(t *testing.T) {
	// GIVEN
	service := NewMockService()
	api := NewStatusApi(service)
	userId := uuid.New()
	status1 := statuses.Status{Id: uuid.New(), Content: "test status 1", UserId: userId}
	status2 := statuses.Status{Id: uuid.New(), Content: "test status 2", UserId: userId}
	service.statuses = []statuses.Status{status1, status2}

	router := chi.NewRouter()
	api.ConfigureRouter(router)

	req, err := http.NewRequest(http.MethodGet, "/users/"+userId.String()+"/statuses", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	// WHEN
	router.ServeHTTP(rr, req)

	// THEN
	assert.Equal(t, http.StatusOK, rr.Code)

	var statusesResponse statuses.StatusesResponse
	err = json.NewDecoder(rr.Body).Decode(&statusesResponse)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(statusesResponse.Statuses))
	assert.Equal(t, status1.Id, statusesResponse.Statuses[0].Id)
	assert.Equal(t, status1.Content, statusesResponse.Statuses[0].Content)
	assert.Equal(t, status1.UserId, statusesResponse.Statuses[0].UserId)
	assert.Equal(t, status2.Id, statusesResponse.Statuses[1].Id)
	assert.Equal(t, status2.Content, statusesResponse.Statuses[1].Content)
	assert.Equal(t, status2.UserId, statusesResponse.Statuses[1].UserId)
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
	createStatusRequest := statuses.CreateStatusRequest{Content: status.Content}
	requestBody, err := json.Marshal(createStatusRequest)
	assert.NoError(t, err)

	router := chi.NewRouter()
	api.ConfigureRouter(router)

	req, err := http.NewRequest(http.MethodPost, "/statuses", bytes.NewReader(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-user", status.UserId.String())

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
