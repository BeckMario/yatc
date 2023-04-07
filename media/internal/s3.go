package media

import (
	"context"
	"encoding/json"
	"github.com/dapr/go-sdk/client"
	"yatc/internal"
)

type S3 interface {
	Create(data []byte, key string) error
	Upload(path string, key string) error
	Presign(ttl string, key string) (string, error)
}

type DaprS3 struct {
	client client.Client
	config internal.S3BindingConfig
}

func NewDaprS3(client client.Client, config internal.S3BindingConfig) *DaprS3 {
	return &DaprS3{client, config}
}

func (s3 *DaprS3) Create(data []byte, key string) error {
	invokeBindingRequest := client.InvokeBindingRequest{
		Name:      s3.config.Name,
		Operation: "create",
		Data:      data,
		Metadata:  map[string]string{"key": key},
	}
	return s3.client.InvokeOutputBinding(context.Background(), &invokeBindingRequest)
}

func (s3 *DaprS3) Upload(path string, key string) error {
	invokeBindingRequest := client.InvokeBindingRequest{
		Name:      s3.config.Name,
		Operation: "create",
		Data:      make([]byte, 0),
		Metadata:  map[string]string{"key": key, "filePath": path},
	}
	return s3.client.InvokeOutputBinding(context.Background(), &invokeBindingRequest)
}

func (s3 *DaprS3) Presign(ttl string, key string) (string, error) {
	invokeBindingRequest := client.InvokeBindingRequest{
		Name:      s3.config.Name,
		Operation: "presign",
		Data:      nil,
		Metadata:  map[string]string{"key": key, "presignTTL": ttl},
	}

	response, err := s3.client.InvokeBinding(context.Background(), &invokeBindingRequest)
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
