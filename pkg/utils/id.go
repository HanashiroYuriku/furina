package utils

import (
	"strings"
	"github.com/google/uuid"
)

func GenerateID(prefix string) string {
	newUUID := uuid.New().String()

	cleanUUID := strings.ReplaceAll(newUUID, "-", "")

	if prefix != "" {
		return strings.ToUpper(prefix) + "-" + cleanUUID
	}
	return "ID-" + cleanUUID
}