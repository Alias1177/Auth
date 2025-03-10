package main

import "Auth/pkg/logger"

func main() {
	log, err := logger.NewSimpleLogger("info")
}
