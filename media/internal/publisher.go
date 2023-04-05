package media

import (
	"context"
	"github.com/dapr/go-sdk/client"
	"yatc/internal"
)

type Publisher interface {
	Publish(mediaId string) error
}

type DaprMediaPublisher struct {
	client client.Client
	config internal.PubSubConfig
}

func NewDaprMediaPublisher(client client.Client, config internal.PubSubConfig) *DaprMediaPublisher {
	return &DaprMediaPublisher{client, config}
}

func (pub *DaprMediaPublisher) Publish(mediaId string) error {
	return pub.client.
		PublishEvent(context.Background(), pub.config.Name, pub.config.Topic, struct{ MediaId string }{mediaId})
}
