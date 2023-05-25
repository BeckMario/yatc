package statuses

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"io"
	"net/http"
	"yatc/internal"
)

const (
	BaseRoute = "/internal/pubsub/receive"
)

type Subscriber interface {
	Subscribe(handler func(ctx context.Context, status Status))
}

type StatusCloudEvent struct {
	Id     string `json:"id"`
	Status Status `json:"data"`
}

type DaprStatusSubscriber struct {
	router chi.Router
	logger *zap.Logger
	route  string
}

func getSubscribeHandler(config internal.PubSubConfig, route string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		subscriptions := []struct {
			PubSubName string `json:"pubsubname"`
			Topic      string `json:"topic"`
			Routes     string `json:"route"`
		}{
			{
				config.Name,
				config.Topic,
				route,
			},
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, subscriptions)
	}
}

func NewDaprStatusSubscriber(router chi.Router, logger *zap.Logger, config internal.PubSubConfig) *DaprStatusSubscriber {
	route := fmt.Sprintf("%s/%s", BaseRoute, config.Topic)
	router.Get("/dapr/subscribe", getSubscribeHandler(config, route))

	return &DaprStatusSubscriber{router, logger, route}
}

// Subscribe Currently there can only be one subscribe handler
func (sub *DaprStatusSubscriber) Subscribe(handler func(ctx context.Context, status Status)) {
	sub.router.Post(sub.route, func(w http.ResponseWriter, r *http.Request) {
		//TODO: Do this in middleware of router
		trace := r.Header.Get("Traceparent")
		ctx := context.Background()
		if trace != "" {
			ctx = context.WithValue(ctx, internal.ContextKeyTraceParent, trace)
		}
		cloudEvent := &StatusCloudEvent{}
		var bodyBytes []byte
		bodyBytes, _ = io.ReadAll(r.Body)
		err := json.Unmarshal(bodyBytes, &cloudEvent)
		if err != nil {
			// Shouldn't normally happen when using dapr to publish and subscribe
			sub.logger.DPanic("message not a cloudevent", zap.Error(err))
		}
		handler(ctx, cloudEvent.Status)
		render.Status(r, http.StatusOK)
	})
}
