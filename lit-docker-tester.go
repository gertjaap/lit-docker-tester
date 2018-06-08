package main

import (
	"log"
	"os"

	"github.com/gertjaap/lit-docker-tester/testscripts"
)

func handleErrorIfNeeded(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(0)
	}
}

func main() {
	testscripts.MultihopTest()
}
