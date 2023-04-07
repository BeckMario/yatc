package media

import (
	"bytes"
	"fmt"
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"io"
	"net/http"
	"strings"
	"yatc/internal"
)

type Api struct {
	service Service
}

var (
	allowedContentTypes = []string{"image/gif", "image/jpeg", "image/png", "video/mp4"}
)

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
	var media openapi_types.File
	_, header, err := r.FormFile("media")
	if err != nil {
		return nil, err
	}

	media.InitFromMultipart(header)
	return &MediaUpload{media}, nil
}

func (api *Api) UploadMedia(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1_000_000) // 1Mb
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusBadRequest)
		return
	}

	mediaUpload, err := NewMediaUpload(r)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusBadRequest)
		return
	}

	reader, err := mediaUpload.Media.Reader()
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		return
	}

	reader, contentType, err := getContentType(io.Reader(reader))
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		return
	}

	if !isContentTypeAllowed(contentType) {
		internal.ReplyWithError(w, r, fmt.Errorf("content type not allowed: %s", contentType), http.StatusBadRequest)
		return
	}

	_, extension, _ := strings.Cut(contentType, "/")

	media := &Media{
		metadata: Metadata{
			Extension: extension,
		},
		fileName: mediaUpload.Media.Filename(),
		reader:   io.Reader(reader),
	}

	mediaId, err := api.service.UploadFile(media)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		return
	}

	render.JSON(w, r, MediaUploadResponse{MediaId: mediaId})
}

func getContentType(reader io.Reader) (io.ReadCloser, string, error) {
	mimeSniffReader := io.LimitReader(reader, 512)
	mimeSniff, err := io.ReadAll(mimeSniffReader)
	if err != nil {
		return nil, "", err
	}
	return io.NopCloser(io.MultiReader(bytes.NewReader(mimeSniff), reader)),
		http.DetectContentType(mimeSniff), nil
}

func isContentTypeAllowed(contentType string) bool {
	for _, allowedContentType := range allowedContentTypes {
		if contentType == allowedContentType {
			return true
		}
	}
	return false
}

func (api *Api) DownloadMedia(w http.ResponseWriter, r *http.Request, mediaId string) {
	url, err := api.service.DownloadFile(mediaId)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Add("Location", url)
	w.WriteHeader(http.StatusSeeOther)
}
