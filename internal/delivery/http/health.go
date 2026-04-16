package http

import (
	"be-ayaka/config"
	ayaka "be-ayaka/pkg/logger"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type HealthCheckHandler struct {
	Config *config.Config
	db *gorm.DB
}

func NewHealthCheckHandler(cfg *config.Config, db *gorm.DB) *HealthCheckHandler {
	return &HealthCheckHandler{
		Config: cfg,
		db: db,
	}
}

func (h *HealthCheckHandler) Check(c *fiber.Ctx) error {
	go ayaka.Log("SYSTEM", "INFO", "Ayaka Server is running")
	appStatus := "UP"
	dbStatus := "UP"
	httpCode := fiber.StatusOK

	sqlDB, err := h.db.DB() 
	
	if err != nil || sqlDB.Ping() != nil {
		appStatus = "DOWN"
		dbStatus = "DOWN"
		httpCode = fiber.StatusServiceUnavailable
	}

	return c.Status(httpCode).JSON(fiber.Map{
		"name":    h.Config.App.Name,
		"status":  appStatus,
		"time":    time.Now().Format(time.RFC3339),
		"components": fiber.Map{
			"server":   "UP",       
			"database": dbStatus,   
		},
	})
}
