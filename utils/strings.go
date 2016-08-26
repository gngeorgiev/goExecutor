package utils

import (
	"strings"

	"github.com/satori/go.uuid"
)

func RandomString() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}
