package internal

import (
	"errors"
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
