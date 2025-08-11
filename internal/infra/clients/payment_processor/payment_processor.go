package paymentprocessor

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/marincor/rinha-de-backend-2025-marincor-golang/constants"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/helpers"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/entities"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/infra/clients/request"
)

var (
	errInvalidStatusCode = errors.New("invalid status code")
	errHealthFailing     = errors.New("health check failing")
)

const (
	throttlingFactor = 5.1 * float64(time.Second)
	maxRequestTime   = 150 * time.Millisecond

	maxRetries   = 10
	initialDelay = time.Millisecond
	multiplier   = 3
	randomInt    = 4
)

type Client struct {
	baseURL           string
	request           *request.HTTPRequest
	processorProvider entities.ProcessorProvider
	failing           bool
}

func New(baseURL string, processorProvider entities.ProcessorProvider) *Client {
	client := &Client{
		baseURL:           baseURL,
		request:           request.New(),
		processorProvider: processorProvider,
		failing:           false,
	}

	if processorProvider == entities.Default {
		go func(currentClient *Client) {
			init := false

			healthRequestClient := request.New()

			healthTimeout := 30

			healthRequestClient.SetNewTimeout(time.Duration(healthTimeout) * time.Second)

			for {
				if init {
					time.Sleep(time.Duration(throttlingFactor))
				}

				init = true

				health, err := currentClient.health(healthRequestClient)
				if err != nil {
					go log.Print(
						map[string]interface{}{
							"message": "error getting payments health",
							"error":   err,
						},
					)

					currentClient.failing = false

					continue
				}

				if health.Failing {
					go log.Print(
						map[string]interface{}{
							"message":         "health check failing",
							"error":           errHealthFailing,
							"failing":         health.Failing,
							"minResponseTime": health.MinResponseTime,
						},
					)

					currentClient.failing = true
				} else {
					currentClient.failing = false
				}
			}
		}(client)
	}

	return client
}

func (c *Client) ProcessPayment(paymentRequest *entities.PaymentRequest) (*entities.PaymentResponse, error) {
	if c.failing {
		return nil, errHealthFailing
	}

	body := map[string]any{
		"correlationId": paymentRequest.CorrelationID,
		"amount":        paymentRequest.Amount,
		"requestedAt":   paymentRequest.RequestedAt,
	}

	headers := map[string]string{}

	response, err := helpers.ExponentialBackoffRetry(func() (*request.Response, error) {
		response, err := c.request.POST(c.baseURL+"/payments", headers, body)
		if err != nil {
			return response, fmt.Errorf("error processing payment: %w", err)
		}

		if response.StatusCode != constants.HTTPStatusOK {
			return response, constants.NewErrorWrapper(errInvalidStatusCode, fmt.Sprintf("Error to process payment: %s", response.Status))
		}

		return response, nil
	}, maxRetries, initialDelay, multiplier, randomInt)
	if err != nil {
		return nil, fmt.Errorf("error processing payment: %w", err)
	}

	if response.StatusCode != constants.HTTPStatusOK {
		return &entities.PaymentResponse{
			Message:           "error invalid status",
			ProcessorProvider: c.processorProvider,
		}, fmt.Errorf("error to process payment: %w: %s", constants.ErrInvalidStatusCode, response.Status)
	}

	return &entities.PaymentResponse{
		Message:           "success",
		ProcessorProvider: c.processorProvider,
	}, nil
}

func (c *Client) PaymentsSummary(filters *entities.PaymentSummaryFilters) (*entities.PaymentSummaryResponse, error) {
	endpointURL, err := url.Parse(c.baseURL + "/admin/payments-summary")
	if err != nil {
		return nil, fmt.Errorf("parsing endpoint url: %w", err)
	}

	if filters != nil {
		query := endpointURL.Query()

		if filters.From != nil {
			query.Add("from", filters.From.Format(constants.DefaultTimeFormat))
		}

		if filters.To != nil {
			query.Add("to", filters.To.Format(constants.DefaultTimeFormat))
		}

		endpointURL.RawQuery = query.Encode()
	}

	headers := map[string]string{}

	response, err := c.request.GET(endpointURL.String(), headers)
	if err != nil {
		return nil, fmt.Errorf("error getting payments summary: %w", err)
	}

	if response.StatusCode != constants.HTTPStatusOK {
		return nil, constants.NewErrorWrapper(errInvalidStatusCode, fmt.Sprintf("Error to get summary: %s", response.Status))
	}

	var paymentSummaryResponse PaymentSummaryResponse
	err = helpers.Unmarshal(response.Body, &paymentSummaryResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling payment summary response: %w", err)
	}

	return &entities.PaymentSummaryResponse{
		Default: entities.Summary{
			TotalRequests: paymentSummaryResponse.Default.TotalRequests,
			TotalAmount:   paymentSummaryResponse.Default.TotalAmount,
		},
		Fallback: entities.Summary{
			TotalRequests: paymentSummaryResponse.Fallback.TotalRequests,
			TotalAmount:   paymentSummaryResponse.Fallback.TotalAmount,
		},
	}, nil
}

func (c *Client) health(healthRequestClient *request.HTTPRequest) (*Health, error) {
	headers := map[string]string{}

	response, err := healthRequestClient.GET(c.baseURL+"/payments/service-health", headers)
	if err != nil {
		return nil, fmt.Errorf("error getting payments health: %w", err)
	}

	if response.StatusCode != constants.HTTPStatusOK {
		return nil, constants.NewErrorWrapper(errInvalidStatusCode, response.Status)
	}

	var paymentHealthResponse Health
	err = helpers.Unmarshal(response.Body, &paymentHealthResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling payment health response: %w", err)
	}

	return &paymentHealthResponse, nil
}
