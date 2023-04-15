package statuses

import (
	"errors"
	"fmt"
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

func StatusFromCreateStatusRequest(request statuses.CreateStatusRequest) statuses.Status {
	return statuses.Status{
		Id:      uuid.New(),
		Content: request.Content,
		UserId:  request.UserId,
	}
}

func NewStatusApi(service statuses.Service) *Api {
	return &Api{service: service}
}

func (api *Api) ConfigureRouter(router chi.Router, middlewares ...func(http.Handler) http.Handler) {
	var middlewareFuncs []statuses.MiddlewareFunc
	for _, middleware := range middlewares {
		middlewareFuncs = append(middlewareFuncs, middleware)
	}

	handler := statuses.HandlerWithOptions(api,
		statuses.ChiServerOptions{
			ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				internal.ReplyWithError(w, r, err, http.StatusBadRequest)
			},
			Middlewares: middlewareFuncs,
		})

	router.Mount("/", handler)
}

func (api *Api) GetStatuses(w http.ResponseWriter, r *http.Request) {
	service := api.service
	allStatuses, err := service.GetStatuses()
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		return
	}

	statusResponses := make([]statuses.StatusResponse, len(allStatuses))
	for i, status := range allStatuses {
		statusResponses[i] = statuses.StatusResponseFromStatus(status)
	}

	internal.ReplyWithStatusOkWithJSON(w, r, statusResponses)
}

func (api *Api) CreateStatus(w http.ResponseWriter, r *http.Request) {
	service := api.service
	var createStatusRequest statuses.CreateStatusRequest
	err := render.Decode(r, &createStatusRequest)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusBadRequest)
		return
	}

	status := StatusFromCreateStatusRequest(createStatusRequest)
	status, err = service.CreateStatus(status)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusInternalServerError)
		return
	}

	internal.ReplyWithStatusWithJSON(w, r, http.StatusCreated, statuses.StatusResponseFromStatus(status))
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
	userId := r.Context().Value("userId")

	fmt.Println("HERE")
	fmt.Println(userId)
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
