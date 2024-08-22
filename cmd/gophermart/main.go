package main

import (
	"fmt"
	"log"
)

const (
	RunAddress = "127.0.0.1:8080"
)

func main() {
	config, err := NewConfig()
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to get config. Error:", err)
	}

	// DEBUG
	fmt.Print(config)
}
