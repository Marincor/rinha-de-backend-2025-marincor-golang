package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/infra/app/appinstance"
)

func route() *fiber.App {
	// middlewares
	appinstance.Data.Server.Use(logger.New())
	appinstance.Data.Server.Use(recover.New())

	appinstance.Data.Server.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	healthController := makeHealthController()

	healthGroup := appinstance.Data.Server.Group("/health")
	healthGroup.Get("", healthController.Check).Name("health_check")

	return appinstance.Data.Server
}
