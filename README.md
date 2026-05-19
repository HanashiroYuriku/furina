# Furina (Genshin Impact Material Tracer API)

A lightweight, scalable backend service to track Genshin Impact character ascensions, weapon upgrades, and daily farming schedules. 

## Tech Stack & Architecture

This project is built on top of my personal enterprise-grade boilerplate, the **[be-ayaka template](https://github.com/HanashiroYuriku/be-ayaka)**. 

It implements a strict **Hexagonal Architecture (Ports & Adapters)** utilizing **Golang, Fiber v2, PostgreSQL, and GORM**. This structure ensures high testability, easy maintenance, and a clean separation of business logic from delivery mechanisms.

## Features

* **Daily Domain Tracker:** Check which domains and materials are available today.
* **Ascension Calculator:** Calculate total required materials (Mora, Boss Drops, Books) for specific characters/weapons.
* **Inventory Management:** Track current in-game items and calculate shortages.

## 📂 Project Structure
```text
furina/
├── cmd/                   # Application entry point (CLI commands, root.go, server.go)
├── config/                # Configuration setup (Viper, Godotenv)
├── internal/              # Private application codebase
│   ├── adapter/           # Infrastructure layer (Database connections, 3rd party APIs)
│   │   ├── database/      # Database initialization & driver setup (PostgreSQL)
│   │   ├── email/         # SMTP implementation for system notifications
│   │   └── repository/    # GORM implementations (Fulfills core repository contracts)
│   ├── bootstrap/         # The Wiring (Dependency Injection, App Init, Routes)
│   ├── core/              # Core Business Logic (The Holy Grail - Framework Agnostic)
│   │   ├── customerrors/  # Standardized Domain & Application Custom Errors
│   │   ├── entity/        # Business Rules: Pure Data Structs
│   │   ├── port/          # Contracts: Inbound (Service) & Outbound (Repository) Interfaces
│   │   └── service/       # Application Business Rules: Use Cases / Orchestrator
│   ├── delivery/          # Transport Mechanism (The Receptionist)
│   │   └── http/          # HTTP Handlers (Fiber controllers: health, user, auth)
│   ├── middleware/        # HTTP Interceptors (JWT Auth, Role checking, Request ID)
│   └── testingutils/      # Test fixtures, mock generators, and test suites helpers
├── logs/                  # Application runtime and error log files
├── pkg/                   # Reusable, domain-agnostic utilities (Hash, JWT, Logger, Validator)
├── .env.sample            # Local environment variables template
├── Dockerfile             # Multi-stage production Docker build recipe
├── config.yaml            # Configuration mapping with environment variable interpolation
├── TESTING.md             # Guide for running unit tests, integration tests, and generating coverage
└── main.go                # Application root executing the cmd command processor
```

## Getting Started
1. Requirements
Go 1.25.1 or higher

PostgreSQL installed and running

2. Setup
Clone the repository and prepare your database (e.g., furina_db). Then, configure your environment variables:
```bash
cp .env.sample .env
```
_(Update the .env file with your local database credentials)_

3. Run the App
Install dependencies and start the server. The database auto-migration will run automatically.
``` bash
go mod tidy
go run cmd/api/main.go
```
The API is now running at http://localhost:8000.
Check system status: GET http://localhost:8000/health

## Feature Development Workflow
To maintain architecture consistency, new features are built inside-out:

1. `core/entity` (Data Struct) -> core/repository (Interface)

2. `adapter/repository` (GORM Implementation)

3. `core/service` (Business Logic)

4. `delivery/http` (Endpoint Handler)

5. Wire everything in `bootstrap/builder.go` & `bootstrap/routes.go`