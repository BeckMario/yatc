package media

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"io"
)

type Metadata struct {
	Extension string
}

type Media struct {
	metadata Metadata
	fileName string
	reader   *io.Reader
}

type Service interface {
	UploadFile(media *Media) (string, error)
	DownloadFile(mediaId string) (string, error)
}

type MediaService struct {
	s3 S3
}

func NewMediaService(s3 S3) *MediaService {
	return &MediaService{s3}
}

func (service *MediaService) UploadFile(media *Media) (string, error) {
	mediaId := uuid.New()
	key := fmt.Sprintf("%s.%s", mediaId, media.metadata.Extension)

	//TODO: Save file to temporary location and use invoke binding request with file path. so the file doesnt get fully read into memory
	// -> shared volume for dapr sidecar and app needed
	reader := bufio.NewReader(*media.reader)
	mediaBytes, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	b64String := base64.StdEncoding.EncodeToString(mediaBytes)
	data := []byte(b64String)

	err = service.s3.Create(data, key)
	if err != nil {
		return "", err
	}

	return key, nil
}

func (service *MediaService) DownloadFile(mediaId string) (string, error) {
	return service.s3.Presign("5m", mediaId)
}
