package testscripts

import (
	"log"
	"os"
	"strings"
)

func handleErrorIfNeeded(err error) {
	if err != nil {
		if strings.Index(err.Error(), "address already in use") == -1 {
			log.Fatal("Fatal error occurred in test: ", err.Error())
			os.Exit(0)
		}
	}
}
