package paymentprocessor

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/marincor/rinha-de-backend-2025-marincor-golang/constants"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/helpers"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/domain/entities"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/internal/infra/clients/request"
)

var errInvalidStatusCode = errors.New("invalid status code")

type Client struct {
	baseURL           string
	request           *request.HTTPRequest
	processorProvider entities.ProcessorProvider
}

func New(baseURL string, processorProvider entities.ProcessorProvider) *Client {
	return &Client{
		baseURL:           baseURL,
		request:           request.New(),
		processorProvider: processorProvider,
	}
}

func (c *Client) ProcessPayment(paymentRequest *entities.PaymentRequest) (*entities.PaymentResponse, error) {
	body := map[string]any{
		"correlationId": paymentRequest.CorrelationID,
		"amount":        paymentRequest.Amount,
		"requestedAt":   paymentRequest.RequestedAt,
	}

	headers := map[string]string{}

	response, err := c.request.POST(c.baseURL+"/payments", headers, body)
	if err != nil {
		return nil, fmt.Errorf("error processing payment: %w", err)
	}

	if response.StatusCode != constants.HTTPStatusOK {
		return &entities.PaymentResponse{
			Message:           "error invalid status",
			ProcessorProvider: c.processorProvider,
		}, fmt.Errorf("%w: %s", constants.ErrInvalidStatusCode, response.Status)
	}

	var paymentResponse Response
	err = helpers.Unmarshal(response.Body, &paymentResponse)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling payment response: %w", err)
	}

	return &entities.PaymentResponse{
		Message:           paymentResponse.Message,
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
		return nil, constants.NewErrorWrapper(errInvalidStatusCode, response.Status)
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
