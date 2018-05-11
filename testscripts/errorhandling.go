package testscripts

import (
	"log"
	"os"
)

func handleErrorIfNeeded(err error) {
	if err != nil {

		log.Fatal("Fatal error occurred in test: ", err.Error())
		os.Exit(0)
	}
}
