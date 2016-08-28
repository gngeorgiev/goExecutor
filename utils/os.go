package utils

import (
	"log"
	"os"
)

var cwd string

func GetWd() string {
	if cwd != "" {
		return cwd
	}

	wd, wdError := os.Getwd()
	if wdError != nil {
		log.Fatal(wdError)
	}

	cwd = wd
	return wd
}
