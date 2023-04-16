package statuses

import (
	"errors"
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"net/http"
	"yatc/internal"
	statuses "yatc/status/pkg"
)

type Api struct {
	service statuses.Service
}

func StatusFromCreateStatusRequest(request statuses.CreateStatusRequest, userId uuid.UUID) statuses.Status {
	return statuses.Status{
		Id:      uuid.New(),
		Content: request.Content,
		UserId:  userId,
	}
}

func NewStatusApi(service statuses.Service) *Api {
	return &Api{service: service}
}

func (api *Api) ConfigureRouter(router chi.Router) {
	handler := statuses.HandlerWithOptions(api,
		statuses.ChiServerOptions{
			ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				internal.ReplyWithError(w, r, err, http.StatusBadRequest)
			},
		})

	router.Mount("/", handler)
}

func (api *Api) CreateStatus(w http.ResponseWriter, r *http.Request, params statuses.CreateStatusParams) {
	service := api.service
	var createStatusRequest statuses.CreateStatusRequest
	err := render.Decode(r, &createStatusRequest)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusBadRequest)
		return
	}

	status := StatusFromCreateStatusRequest(createStatusRequest, params.XUser)
	status, err = service.CreateStatus(status)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		return
	}

	internal.ReplyWithStatusWithJSON(w, r, http.StatusCreated, statuses.StatusResponseFromStatus(status))
}

func (api *Api) GetStatuses(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID) {
	service := api.service
	allStatuses, err := service.GetStatuses(userId)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		return
	}

	statusResponses := make([]statuses.StatusResponse, len(allStatuses))
	for i, status := range allStatuses {
		statusResponses[i] = statuses.StatusResponseFromStatus(status)
	}

	internal.ReplyWithStatusOkWithJSON(w, r, statuses.StatusesResponse{Statuses: statusResponses})
}

func (api *Api) DeleteStatus(w http.ResponseWriter, r *http.Request, statusId uuid.UUID) {
	status, err := api.service.DeleteStatus(statusId)
	if err != nil {
		if errors.Is(err, internal.NotFoundError(statusId)) {
			internal.ReplyWithError(w, r, err, http.StatusNotFound)
		} else {
			internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	internal.ReplyWithStatusOkWithJSON(w, r, statuses.StatusResponseFromStatus(status))
}

func (api *Api) GetStatus(w http.ResponseWriter, r *http.Request, statusId uuid.UUID) {
	status, err := api.service.GetStatus(statusId)
	if err != nil {
		if errors.Is(err, internal.NotFoundError(statusId)) {
			internal.ReplyWithError(w, r, err, http.StatusNotFound)
		} else {
			internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		}
		return
	}

	internal.ReplyWithStatusOkWithJSON(w, r, statuses.StatusResponseFromStatus(status))
}
