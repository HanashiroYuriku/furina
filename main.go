package main

import (
	"log"

	"be-ayaka/cmd"
)

// @title Furina (Material Tracer) API
// @version 1.0
// @description API Documentation for Material Tracer - FURINA.
// @contact.name Hanashiro Yuriku
// @host localhost:8000
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("Fail run application: %v", err)
	}
}
