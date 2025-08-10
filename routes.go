package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/contracts"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/infra/app/appinstance"
)

func route(workerPool contracts.WorkerPoolManager) *fiber.App {
	// middlewares
	appinstance.Data.Server.Use(logger.New())
	appinstance.Data.Server.Use(recover.New())

	appinstance.Data.Server.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	healthController := makeHealthController()
	paymentController := makePaymentController(appinstance.Data.Config, workerPool)

	healthGroup := appinstance.Data.Server.Group("/health")
	healthGroup.Get("", healthController.Check).Name("health_check")

	paymentGroup := appinstance.Data.Server.Group("/payments")
	paymentGroup.Post("", paymentController.ProcessPayment).Name("process_payment")

	paymentsSummaryGroup := appinstance.Data.Server.Group("/payments-summary")
	paymentsSummaryGroup.Get("", paymentController.RetrievePaymentSummary).Name("retrieve_payment_summary")

	return appinstance.Data.Server
}
