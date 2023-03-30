package statuses

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"io"
	"net/http"
)

const (
	PubSubName = "pubsub"
	Topic      = "status"
	BaseRoute  = "/internal/pubsub/receive"
)

var route = fmt.Sprintf("%s/%s", BaseRoute, Topic)

type Subscriber interface {
	Subscribe(handler func(status Status)) error
}

type StatusCloudEvent struct {
	Id     string `json:"id"`
	Status Status `json:"data"`
}

type DaprStatusSubscriber struct {
	router chi.Router
	logger *zap.Logger
}

func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	subscriptions := []struct {
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

	render.Status(r, http.StatusOK)
	render.JSON(w, r, subscriptions)
}

func NewDaprTweetSubscriber(router chi.Router, logger *zap.Logger) *DaprStatusSubscriber {
	router.Get("/dapr/subscribe", subscribeHandler)
	return &DaprStatusSubscriber{router, logger}
}

// Subscribe Currently there can only be one subscribe handler
func (sub *DaprStatusSubscriber) Subscribe(handler func(status Status)) {
	route := fmt.Sprintf("%s/%s", BaseRoute, Topic)
	sub.router.Post(route, func(w http.ResponseWriter, r *http.Request) {
		cloudEvent := &StatusCloudEvent{}
		var bodyBytes []byte
		bodyBytes, _ = io.ReadAll(r.Body)
		err := json.Unmarshal(bodyBytes, &cloudEvent)
		if err != nil {
			// Shouldn't normally happen when using dapr to publish and subscribe
			sub.logger.DPanic("message not a cloudevent", zap.Error(err))
		}
		handler(cloudEvent.Status)
		render.Status(r, http.StatusOK)
	})
}
