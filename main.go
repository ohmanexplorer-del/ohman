package main

import (
	"log"

	"ohman/app"
)

func main() {
	log.Println("Starting Ohman Explorer...")

	a, err := app.New(".")
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	a.Run()
}
