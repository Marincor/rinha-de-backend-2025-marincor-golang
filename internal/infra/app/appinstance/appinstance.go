package appinstance

import (
	"github.com/gofiber/fiber/v2"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/config"
)

type Application struct {
	Config *config.Config
	Server *fiber.App
}

var Data *Application
