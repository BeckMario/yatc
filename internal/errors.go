package internal

import (
	"fmt"
	"github.com/google/uuid"
)

type NotFoundError uuid.UUID

func (err NotFoundError) Error() string {
	return fmt.Sprintf("entity with uuid {%s} not found", uuid.UUID(err).String())
}
