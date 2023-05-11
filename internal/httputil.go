package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type ErrorResponse struct {
	Method    string
	Path      string
	Timestamp time.Time
	Message   string
}

func ReplyWithError(w http.ResponseWriter, r *http.Request, err error, status int) {
	lvl := zap.InfoLevel
	if status == 500 {
		lvl = zap.ErrorLevel
	}
	// Global logger because to lazy to do it right
	zap.L().Log(lvl, "Error while serving http route", zap.String("error", err.Error()), zap.Int("status", status))

	var notFound NotFoundError
	msg := "Error"
	if errors.As(err, &notFound) {
		msg = uuid.UUID(notFound).String()
	}

	errorRes := ErrorResponse{
		Method:    r.Method,
		Path:      r.RequestURI,
		Timestamp: time.Now().UTC(),
		Message:   msg,
	}
	ReplyWithStatusWithJSON(w, r, status, errorRes)
}

func ReplyWithStatusOkWithJSON(w http.ResponseWriter, r *http.Request, json interface{}) {
	ReplyWithStatusWithJSON(w, r, http.StatusOK, json)
}

func ReplyWithStatusWithJSON(w http.ResponseWriter, r *http.Request, status int, json interface{}) {
	render.Status(r, status)
	render.JSON(w, r, json)
}

func ReplyWithStatusOk(r *http.Request) {
	render.Status(r, http.StatusOK)
}

func ToClientError(response *http.Response, err error) *ClientError {
	if err != nil {
		return NewClientError(nil, err)
	}

	if response.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		err := render.DecodeJSON(response.Body, &errorResponse)
		if err != nil {
			return NewClientError(nil, err)
		}

		if response.StatusCode == http.StatusNotFound {
			id, err := uuid.Parse(errorResponse.Message)
			if err != nil {
				return NewClientError(&errorResponse, err)
			}
			return NewClientError(&errorResponse, NotFoundError(id))
		}

		return NewClientError(&errorResponse, nil)
	}
	return nil
}

func OapiClientAuthRequestFn() func(ctx context.Context, req *http.Request) error {
	return func(ctx context.Context, req *http.Request) error {
		value := ctx.Value(ContextKeyAuthorization)
		if value != nil {
			jwt, ok := value.(string)
			if ok {
				req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwt))
			}
		}
		return nil
	}
}

func OapiClientTraceRequestFn() func(ctx context.Context, req *http.Request) error {
	return func(ctx context.Context, req *http.Request) error {
		value := ctx.Value(ContextKeyTraceParent)
		if value != nil {
			trace, ok := value.(string)
			if ok {
				req.Header.Add("Traceparent", trace)
			}
		}
		return nil
	}
}
