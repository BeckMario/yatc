package media

import (
	"encoding/json"
	"fmt"
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"yatc/internal"
)

type Api struct {
	service *Service
}

func NewMediaApi() *Api {
	return &Api{nil}
}

func (api *Api) ConfigureRouter(router chi.Router) {
	handler := HandlerWithOptions(api,
		ChiServerOptions{ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			internal.ReplyWithError(w, r, err, http.StatusBadRequest)
		}})

	router.Mount("/", handler)
}

func NewMediaUpload(r *http.Request) (*MediaUpload, error) {
	metadataJson := r.FormValue("mediaMetadata")
	var mediaMetadata MediaMetadata
	err := json.Unmarshal([]byte(metadataJson), &mediaMetadata)
	if err != nil {
		return nil, err
	}

	var media openapi_types.File
	_, header, err := r.FormFile("media")
	if err != nil {
		return nil, err
	}

	media.InitFromMultipart(header)
	return &MediaUpload{media, mediaMetadata}, nil
}

func NewMediaFromMediaUpload(upload *MediaUpload) (*Media, error) {
	reader, err := upload.Media.Reader()
	if err != nil {
		return nil, err
	}

	metadata := Metadata{Format: upload.MediaMetadata.MediaFormat}

	return &Media{metadata, upload.Media.Filename(), &reader}, nil
}

func (api *Api) UploadMedia(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusBadRequest)
	}

	mediaUpload, err := NewMediaUpload(r)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusBadRequest)
	}

	media, err := NewMediaFromMediaUpload(mediaUpload)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
	}

	fmt.Println(media.metadata)

	join := filepath.Join(".", media.fileName)
	file, err := os.Create(join)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
	}
	defer file.Close()

	_, err = io.Copy(file, *media.reader)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
	}
	reader := *media.reader
	reader.Close()
	//api.service.UploadFile(media)
}

func (api *Api) DownloadMedia(w http.ResponseWriter, r *http.Request, mediaId openapi_types.UUID) {
	//TODO implement me
	panic("implement me")
}
