package statuses

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
	statuses "yatc/status/pkg"
)

type Api struct {
	service statuses.Service
}

type ErrorResponse struct {
	Method    string
	Path      string
	Timestamp time.Time
	Message   string
}

func Error(err error, status int, w http.ResponseWriter, r *http.Request) {
	log.Println(err.Error())
	render.Status(r, status)
	errorRes := ErrorResponse{
		Method:    r.Method,
		Path:      r.RequestURI,
		Timestamp: time.Now().UTC(),
		Message:   "Error",
	}
	render.JSON(w, r, errorRes)
}

func StatusResponseFromStatus(status statuses.Status) StatusResponse {
	return StatusResponse{
		Content: status.Content,
		Id:      status.Id,
		UserId:  status.UserId,
	}
}

func StatusFromCreateStatusRequest(request CreateStatusRequest) statuses.Status {
	return statuses.Status{
		Id:      uuid.New(),
		Content: request.Content,
		UserId:  request.UserId,
	}
}

func NewStatusApi(service statuses.Service) *Api {
	return &Api{service: service}
}

func (api *Api) ConfigureRouter(router chi.Router) {
	handler := HandlerWithOptions(api,
		ChiServerOptions{ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			Error(err, http.StatusBadRequest, w, r)
		}})

	router.Mount("/", handler)
}

func (api *Api) GetStatuses(w http.ResponseWriter, r *http.Request) {
	service := api.service
	allStatuses, err := service.GetStatuses()
	if err != nil {
		Error(err, http.StatusInternalServerError, w, r)
		return
	}

	statusResponses := make([]StatusResponse, len(allStatuses))
	for i, status := range allStatuses {
		statusResponses[i] = StatusResponseFromStatus(status)
	}

	render.JSON(w, r, statusResponses)
}

func (api *Api) CreateStatus(w http.ResponseWriter, r *http.Request) {
	service := api.service
	var createStatusRequest CreateStatusRequest
	err := render.Decode(r, &createStatusRequest)
	if err != nil {
		Error(err, http.StatusBadRequest, w, r)
		return
	}

	status := StatusFromCreateStatusRequest(createStatusRequest)
	status, err = service.CreateStatus(status)
	if err != nil {
		Error(err, http.StatusInternalServerError, w, r)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, StatusResponseFromStatus(status))
}

func (api *Api) DeleteStatus(w http.ResponseWriter, r *http.Request, statusId uuid.UUID) {
	status, err := api.service.DeleteStatus(statusId)
	if err != nil {
		Error(err, http.StatusNotFound, w, r)
		return
	}

	render.JSON(w, r, StatusResponseFromStatus(status))
}

func (api *Api) GetStatus(w http.ResponseWriter, r *http.Request, statusId uuid.UUID) {
	status, err := api.service.GetStatus(statusId)
	if err != nil {
		Error(err, http.StatusNotFound, w, r)
		return
	}

	render.JSON(w, r, StatusResponseFromStatus(status))
}
