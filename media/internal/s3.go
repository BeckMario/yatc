package media

import (
	"context"
	"encoding/json"
	"github.com/dapr/go-sdk/client"
	"yatc/internal"
)

type S3 interface {
	Create(data []byte, key string) error
	Presign(ttl string, key string) (string, error)
}

type DaprS3 struct {
	client client.Client
	config internal.S3BindingConfig
}

func NewDaprS3(client client.Client, config internal.S3BindingConfig) *DaprS3 {
	return &DaprS3{client, config}
}

func (dapr *DaprS3) Create(data []byte, key string) error {
	invokeBindingRequest := client.InvokeBindingRequest{
		Name:      dapr.config.Name,
		Operation: "create",
		Data:      data,
		Metadata:  map[string]string{"key": key},
	}
	return dapr.client.InvokeOutputBinding(context.Background(), &invokeBindingRequest)
}

func (dapr *DaprS3) Presign(ttl string, key string) (string, error) {
	invokeBindingRequest := client.InvokeBindingRequest{
		Name:      dapr.config.Name,
		Operation: "presign",
		Data:      nil,
		Metadata:  map[string]string{"key": key, "presignTTL": ttl},
	}

	response, err := dapr.client.InvokeBinding(context.Background(), &invokeBindingRequest)
	if err != nil {
		return "", err
	}

	data := struct {
		PresignURL string
	}{}

	err = json.Unmarshal(response.Data, &data)
	if err != nil {
		return "", err
	}

	return data.PresignURL, nil
}
