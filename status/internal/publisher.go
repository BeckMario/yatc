package statuses

import (
	"context"
	"github.com/dapr/go-sdk/client"
	statuses "yatc/status/pkg"
)

type Publisher interface {
	Publish(status statuses.Status) error
}

type DaprStatusPublisher struct {
	client client.Client
}

// TODO: in config?
const (
	PubSub = "pubsub"
	Topic  = "status"
)

func NewDaprStatusPublisher(client client.Client) *DaprStatusPublisher {
	return &DaprStatusPublisher{client: client}
}

func (pub *DaprStatusPublisher) Publish(status statuses.Status) error {
	return pub.client.PublishEvent(context.Background(), PubSub, Topic, status)
}
