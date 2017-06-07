package aero

import (
	"github.com/google/uuid"
)

// GenerateUUID generates a unique ID.
func GenerateUUID() string {
	return uuid.New().String()
}
