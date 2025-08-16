package paymentcontroller

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/constants"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/helpers"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/application/dtos"
	processpayment "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/application/usecases/process_payment"
	retrievepaymentsummary "github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/application/usecases/retrieve_payment_summary"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/contracts"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/presentation/controllers/payment/validators"
)

type Controller struct {
	processPaymentUsecase         *processpayment.UseCase
	retrievePaymentSummaryUsecase *retrievepaymentsummary.UseCase
	validator                     *validators.Validator
	workerpool                    contracts.WorkerPoolManager
}

func NewController(
	processPaymentUsecase *processpayment.UseCase,
	retrievePaymentSummaryUsecase *retrievepaymentsummary.UseCase,
	workerpool contracts.WorkerPoolManager,
) *Controller {
	return &Controller{
		processPaymentUsecase:         processPaymentUsecase,
		retrievePaymentSummaryUsecase: retrievePaymentSummaryUsecase,
		validator:                     validators.New(),
		workerpool:                    workerpool,
	}
}

func (c *Controller) ProcessPayment(ctx *fiber.Ctx) error {
	var paymentRequest dtos.PaymentPayload

	if err := helpers.Unmarshal(ctx.Body(), &paymentRequest); err != nil {
		return helpers.CreateResponse(ctx, &helpers.ErrorResponse{
			Message:     "error parsing body",
			Description: err.Error(),
			StatusCode:  constants.HTTPStatusUnprocessableEntity,
		}, constants.HTTPStatusUnprocessableEntity)
	}

	go c.workerpool.Submit(func() {
		_, err := c.processPaymentUsecase.Execute(&paymentRequest)
		if err != nil {
			go log.Print(
				map[string]any{
					"message": "error processing payment",
					"error":   err,
				},
			)
		}
	})

	response := ctx.Response()

	response.Header.Set("Content-Length", "0")

	response.SetStatusCode(constants.HTTPStatusNoContent)

	return nil
}

func (c *Controller) RetrievePaymentSummary(ctx *fiber.Ctx) error {
	var summaryFilters dtos.PaymentSummaryFilters

	if err := ctx.QueryParser(&summaryFilters); err != nil {
		return helpers.CreateResponse(ctx, &helpers.ErrorResponse{
			Message:     "error parsing query params",
			Description: err.Error(),
			StatusCode:  constants.HTTPStatusUnprocessableEntity,
		}, constants.HTTPStatusUnprocessableEntity)
	}

	response, err := c.retrievePaymentSummaryUsecase.Execute(&summaryFilters)
	if err != nil {
		return helpers.CreateResponse(ctx, &helpers.ErrorResponse{
			Message:     "error retrieving payment summary",
			Description: err.Error(),
			StatusCode:  constants.HTTPStatusInternalServerError,
		}, constants.HTTPStatusInternalServerError)
	}

	body, err := helpers.Marshal(response)
	if err != nil {
		return helpers.CreateResponse(ctx, &helpers.ErrorResponse{
			Message:     "error marshalling response",
			Description: err.Error(),
			StatusCode:  constants.HTTPStatusInternalServerError,
		}, constants.HTTPStatusInternalServerError)
	}

	ctx.Response().SetStatusCode(constants.HTTPStatusOK)
	ctx.Response().SetBody(body)

	ctx.Response().Header.Set("Content-Length", "0")

	ctx.Response().Header.SetContentType(fiber.MIMEApplicationJSON)

	return nil
}
