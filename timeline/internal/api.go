package timelines

import (
	"context"
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/go-chi/chi/v5"
	"net/http"
	"yatc/internal"
	statuses "yatc/status/pkg"
	timelines "yatc/timeline/pkg"
)

type Api struct {
	service timelines.Service
}

func NewTimelineApi(service timelines.Service) *Api {
	return &Api{service}
}

func (api *Api) ConfigureRouter(router chi.Router) {
	handler := HandlerWithOptions(api,
		ChiServerOptions{ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			internal.ReplyWithError(w, r, err, http.StatusBadRequest)
		}})
	router.Group(func(r chi.Router) {
		r.Mount("/", handler)
	})
}

func TimelineResponseFromTimeline(timeline timelines.Timeline) TimelineResponse {
	statusResponses := make([]statuses.StatusResponse, len(timeline.Statuses))
	for i, status := range timeline.Statuses {
		statusResponses[i] = statuses.StatusResponseFromStatus(status)
	}

	return TimelineResponse{
		Id:       timeline.UserId,
		Statuses: statusResponses,
	}
}

func (api *Api) GetTimeline(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID) {
	timeline, err := api.service.GetTimeline(context.Background(), userId)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusNotFound)
		return
	}

	internal.ReplyWithStatusOkWithJSON(w, r, TimelineResponseFromTimeline(timeline))
}

func (api *Api) GetTimelineV1(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID) {
	timeline, err := api.service.GetTimeline(context.Background(), userId)
	if err != nil {
		internal.ReplyWithError(w, r, err, http.StatusNotFound)
		return
	}

	internal.ReplyWithStatusOkWithJSON(w, r, TimelineResponseFromTimeline(timeline))
}
