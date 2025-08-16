package app

import (
	"errors"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/config"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/helpers"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/infra/app/appinstance"
)

func ApplicationInit() {
	configs := config.New()

	maxConcurrency := 10_000

	appinstance.Data = &appinstance.Application{
		Config: configs,
		Server: fiber.New(fiber.Config{
			ServerHeader:              "Rinha-Backend-Marincor-2025",
			ErrorHandler:              customErrorHandler,
			JSONEncoder:               helpers.Marshal,
			JSONDecoder:               helpers.Unmarshal,
			DisableKeepalive:          false,
			Prefork:                   false,
			ReduceMemoryUsage:         true,
			Concurrency:               maxConcurrency,
			DisableDefaultDate:        true,
			DisableDefaultContentType: true,
			DisableHeaderNormalizing:  true,
			DisableStartupMessage:     true,
			StrictRouting:             false,
			CaseSensitive:             false,
			UnescapePath:              false,
			CompressedFileSuffix:      ".gz",
		}),
	}
}

func Setup(port string) {
	err := appinstance.Data.Server.Listen(":" + port)

	if errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func customErrorHandler(ctx *fiber.Ctx, err error) error {
	var code int = fiber.StatusInternalServerError
	var capturedError *fiber.Error
	message := "unknown error"

	if errors.As(err, &capturedError) {
		code = capturedError.Code
		if code == fiber.StatusNotFound {
			message = "route not found"
		}
	}

	var errorResponse *helpers.ErrorResponse

	erro := helpers.Unmarshal([]byte(err.Error()), &errorResponse)
	if erro != nil {
		errorResponse = &helpers.ErrorResponse{
			Message:     message,
			StatusCode:  code,
			Description: err.Error(),
		}
	}

	go log.Print(
		map[string]interface{}{
			"message":   message,
			"method":    ctx.Method(),
			"reason":    err.Error(),
			"remote_ip": ctx.IP(),
			"request": map[string]interface{}{
				"query":      ctx.Queries(),
				"url_params": helpers.AllParams(ctx),
			},
		},
	)

	return helpers.CreateResponse(ctx, errorResponse, code) //nolint: wrapcheck
}
