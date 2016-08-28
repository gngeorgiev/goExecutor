package utils

import (
	"fmt"
	"log"
	"time"
)

func TrackTime(start time.Time, format string) {
	log.Println(fmt.Sprintf(format, time.Since(start)))
}
