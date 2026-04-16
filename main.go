package main

import (
	"log"

	"be-ayaka/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("Fail run application: %v", err)
	}
}