package idgenerator

import (
	"github.com/google/uuid"
)

// GenerateID generates a new unique ID using UUID v4
// If prefix is provided, it will be added before the UUID
func GenerateID(prefix ...string) string {
	id := uuid.New().String()
	if len(prefix) > 0 {
		return prefix[0] + "_" + id
	}
	return id
}
