package internal

import (
	"fmt"
	"github.com/google/uuid"
)

type NotFoundError uuid.UUID

func (err NotFoundError) Error() string {
	return fmt.Sprintf("entity with uuid {%s} not found", uuid.UUID(err).String())
}

type ClientError struct {
	cause         error
	errorResponse *ErrorResponse
}

func NewClientError(errorResponse *ErrorResponse, cause error) *ClientError {
	return &ClientError{cause, errorResponse}
}

func (err ClientError) Error() string {
	return fmt.Sprintf("Client Error Cause: %s ErrorResponse: %v", err.cause.Error(), err.errorResponse)
}

func (err ClientError) Unwrap() error {
	return err.cause
}
