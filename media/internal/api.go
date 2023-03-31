package media

import (
	"encoding/json"
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"net/http"
	"yatc/internal"
)

type Api struct {
	service Service
}

func NewMediaApi(service Service) *Api {
	return &Api{service}
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
	//TODO: Check somewhere the media format and allow only a subset e.g. .png, .jpg, .gif, .mp4, etc.

	err := r.ParseMultipartForm(1_000_000) // 1Mb
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

	mediaId, err := api.service.UploadFile(media)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
	}
	render.JSON(w, r, MediaUploadResponse{MediaId: mediaId})
}

func (api *Api) DownloadMedia(w http.ResponseWriter, r *http.Request, mediaId string) {
	url, err := api.service.DownloadFile(mediaId)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
	}
	w.Header().Add("Location", url)
	w.WriteHeader(http.StatusSeeOther)
}
