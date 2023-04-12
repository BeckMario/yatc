package media

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type Metadata struct {
	Extension string
}

type Media struct {
	metadata Metadata
	fileName string
	reader   io.Reader
}

type Service interface {
	UploadFile(media *Media) (string, error)
	DownloadFile(mediaId string, compressed bool) (string, error)
}

type MediaService struct {
	s3        S3
	publisher Publisher
}

func NewMediaService(s3 S3, publisher Publisher) *MediaService {
	return &MediaService{s3, publisher}
}

func (service *MediaService) UploadFile(media *Media) (string, error) {
	mediaId := uuid.New()
	key := fmt.Sprintf("%s.%s", mediaId, media.metadata.Extension)

	file, err := os.Create(fmt.Sprintf("/tmp/tmp-%s", mediaId.String()))
	if err != nil {
		return "", err
	}
	encoder := base64.NewEncoder(base64.StdEncoding, file)

	_, err = io.Copy(encoder, media.reader)
	if err != nil {
		return "", err
	}

	err = encoder.Close()
	if err != nil {
		return "", err
	}

	abs, err := filepath.Abs(file.Name())
	if err != nil {
		return "", err
	}

	err = service.s3.Upload(abs, key)
	if err != nil {
		return "", err
	}

	err = service.publisher.Publish(key)
	if err != nil {
		return "", err
	}

	err = os.Remove(file.Name())
	if err != nil {
		return "", err
	}

	return key, nil
}

func getCompressionExtension(uncompressedExtension string) (string, error) {
	contentType := mime.TypeByExtension(uncompressedExtension)
	imgOrVid, _, _ := strings.Cut(contentType, "/")
	if imgOrVid == "image" {
		return "webp", nil
	} else if imgOrVid == "video" {
		return "webm", nil
	} else {
		return "", errors.New("unrecognized format")
	}
}

func (service *MediaService) DownloadFile(mediaId string, compressed bool) (string, error) {
	id, extension, _ := strings.Cut(mediaId, ".")

	if compressed {
		compressionExtension, err := getCompressionExtension(fmt.Sprintf(".%s", extension))
		if err != nil {
			return "", err
		}
		return service.s3.Presign("5m", fmt.Sprintf("%s.%s", id, compressionExtension))
	} else {
		return service.s3.Presign("5m", mediaId)
	}
}
