package statuses

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDaprStatusSubscriber_SubscribeHandler(t *testing.T) {
	// Given
	req, err := http.NewRequest("GET", "/dapr/subscribe", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/dapr/subscribe", subscribeHandler)

	// When
	r.ServeHTTP(recorder, req)

	// Then
	assert.Equal(t, http.StatusOK, recorder.Code)

	var subscriptions []struct {
		PubSubName string `json:"pubsubname"`
		Topic      string `json:"topic"`
		Routes     string `json:"route"`
	}

	err = json.Unmarshal(recorder.Body.Bytes(), &subscriptions)
	if err != nil {
		t.Fatal(err)
	}

	expected := []struct {
		PubSubName string `json:"pubsubname"`
		Topic      string `json:"topic"`
		Routes     string `json:"route"`
	}{
		{
			PubSubName,
			Topic,
			route,
		},
	}

	assert.Equal(t, expected, subscriptions)
}

func TestDaprStatusSubscriber_Subscribe(t *testing.T) {
	// Given
	router := chi.NewRouter()
	sub := NewDaprTweetSubscriber(router)

	expectedStatus := Status{
		Id:      uuid.New(),
		Content: "Hello world",
		UserId:  uuid.New(),
	}
	handlerCalled := false
	mockHandler := func(status Status) {
		if assert.Equal(t, expectedStatus, status) {
			handlerCalled = true
		}
	}

	// When
	sub.Subscribe(mockHandler)
	event := StatusCloudEvent{
		Id:     uuid.New().String(),
		Status: expectedStatus,
	}
	eventBytes, _ := json.Marshal(event)

	req, err := http.NewRequest("POST", "/internal/pubsub/receive/status", bytes.NewBuffer(eventBytes))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Then
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, handlerCalled, "Handler function should be called with the expected Status")
}