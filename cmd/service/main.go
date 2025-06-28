package main

import (
	"log"

	"github.com/Alias1177/Auth/internal/app"
)

func main() {
	application := app.New()
	if err := application.Run(); err != nil {
		log.Fatal("Failed to run application:", err)
	}
}
