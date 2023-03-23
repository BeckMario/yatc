package statuses

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
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

func NewDaprTweetSubscriber(router chi.Router) *DaprStatusSubscriber {
	router.Get("/dapr/subscribe", subscribeHandler)
	return &DaprStatusSubscriber{router}
}

// Subscribe Currently there can only be one subscribe handler
func (sub *DaprStatusSubscriber) Subscribe(handler func(status Status)) {
	route := fmt.Sprintf("%s/%s", BaseRoute, Topic)
	sub.router.Post(route, func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Receiving Event")

		cloudEvent := &StatusCloudEvent{}
		var bodyBytes []byte
		bodyBytes, _ = io.ReadAll(r.Body)
		err := json.Unmarshal(bodyBytes, &cloudEvent)
		if err != nil {
			fmt.Println(err)
			//TODO: Handle?
			panic("received message is not a cloudevent")
		}

		handler(cloudEvent.Status)
		render.Status(r, http.StatusOK)
	})
}
