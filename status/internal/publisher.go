package statuses

import (
	"context"
	"github.com/dapr/go-sdk/client"
	"yatc/internal"
	statuses "yatc/status/pkg"
)

type Publisher interface {
	Publish(status statuses.Status) error
}

type DaprStatusPublisher struct {
	client client.Client
	config internal.PubSubConfig
}

func NewDaprStatusPublisher(client client.Client, config internal.PubSubConfig) *DaprStatusPublisher {
	return &DaprStatusPublisher{client, config}
}

func (pub *DaprStatusPublisher) Publish(status statuses.Status) error {
	return pub.client.PublishEvent(context.Background(), pub.config.Name, pub.config.Topic, status)
}
