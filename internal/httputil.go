package internal

import (
	"errors"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"log"
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
	log.Println(err.Error())
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
