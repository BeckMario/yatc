package media

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dapr/go-sdk/client"
	"github.com/google/uuid"
	"io"
)

type Metadata struct {
	Format string
}

type Media struct {
	metadata Metadata
	fileName string
	reader   *io.ReadCloser
}

type Service interface {
	UploadFile(media *Media) (string, error)
	DownloadFile(mediaId string) (string, error)
}

type MediaService struct {
	client client.Client
}

func NewMediaService(client client.Client) *MediaService {
	return &MediaService{client}
}

func (service *MediaService) UploadFile(media *Media) (string, error) {
	mediaId := uuid.New()
	key := fmt.Sprintf("%s.%s", mediaId, media.metadata.Format)

	//TODO: Save file to temporary location and use invoke binding request with file path. so the file doesnt get fully read into memory
	// -> shared volume for dapr sidecar and app needed
	reader := bufio.NewReader(*media.reader)
	mediaBytes, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	b64String := base64.StdEncoding.EncodeToString(mediaBytes)
	data := []byte(b64String)

	invokeBindingRequest := client.InvokeBindingRequest{
		Name:      "s3",
		Operation: "create",
		Data:      data,
		Metadata:  map[string]string{"key": key},
	}

	err = service.client.InvokeOutputBinding(context.Background(), &invokeBindingRequest)
	if err != nil {
		return "", err
	}
	return key, nil
}

func (service *MediaService) DownloadFile(mediaId string) (string, error) {
	invokeBindingRequest := client.InvokeBindingRequest{
		Name:      "s3",
		Operation: "presign",
		Data:      nil,
		Metadata:  map[string]string{"key": mediaId, "presignTTL": "5m"},
	}

	response, err := service.client.InvokeBinding(context.Background(), &invokeBindingRequest)
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
