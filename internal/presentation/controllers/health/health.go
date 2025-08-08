package healthcontroller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/constants"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/helpers"
	healthcheck "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/application/usecases/health_check"
)

type Controller struct {
	usecase healthcheck.UseCase
}

func NewController(healthUseCase *healthcheck.UseCase) *Controller {
	return &Controller{
		usecase: *healthUseCase,
	}
}

func (c *Controller) Check(ctx *fiber.Ctx) error {
	check, err := c.usecase.Execute()
	if err != nil {
		return helpers.CreateResponse(ctx, &helpers.ErrorResponse{
			Message:     "error checking health",
			Description: err.Error(),
			StatusCode:  constants.HTTPStatusInternalServerError,
		}, constants.HTTPStatusInternalServerError)
	}

	return helpers.CreateResponse(ctx, check, constants.HTTPStatusOK)
}
