package request

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/marincor/rinha-de-backend-2025-marincor-golang/constants"
	"github.com/marincor/rinha-de-backend-2025-marincor-golang/helpers"
)

type Response struct {
	Status     string
	Body       []byte
	StatusCode int
}

type HTTPRequest struct {
	client  *http.Client
	timeout time.Duration
}

func New() *HTTPRequest {
	return &HTTPRequest{
		client:  &http.Client{},
		timeout: constants.DefaultRequestTimeout,
	}
}

//nolint:cyclop,funlen // long but necessary
func (h *HTTPRequest) request(method, url string, headers map[string]string, rawBody *map[string]any) (*Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	var (
		body           io.Reader
		contentType    string
		bodyBytes      []byte
		responseReturn = &Response{}
	)

	responseReturn.StatusCode = constants.HTTPStatusInternalServerError
	responseReturn.Status = strconv.Itoa(responseReturn.StatusCode)

	contentType = headers["Content-Type"]
	if contentType == "" {
		contentType = "application/json"
	}

	//nolint:nestif // long but necessary
	if rawBody != nil {
		if contentType == "multipart/form-data" {
			newBody, newContentType, err := setFormData(rawBody)
			if err != nil {
				return responseReturn, err
			}

			body = newBody
			contentType = newContentType
		} else {
			bodyBytes, err := helpers.Marshal(*rawBody)
			if err != nil {
				log.Print(
					map[string]interface{}{
						"message": "error marshalling raw body",
						"error":   err,
					},
				)

				return responseReturn, err
			}

			body = bytes.NewBuffer(bodyBytes)
		}
	}

	request, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return responseReturn, err
	}

	for key, value := range headers {
		request.Header.Set(key, value)
	}

	request.Header.Set("Content-Type", contentType)

	response, err := h.client.Do(request)
	if err != nil {
		return responseReturn, err
	}

	bodyBytes, byteErr := io.ReadAll(response.Body)
	if byteErr != nil {
		log.Print(
			map[string]interface{}{
				"message": "error reading response body",
				"error":   byteErr,
			},
		)
	}

	if response.Body != nil {
		defer func() {
			if err := response.Body.Close(); err != nil {
				log.Print(
					map[string]interface{}{
						"message": "error on close response body",
						"error":   err,
					},
				)
			}
		}()
	}

	return &Response{
		Body:       bodyBytes,
		StatusCode: response.StatusCode,
		Status:     response.Status,
	}, err
}

func (h *HTTPRequest) GET(url string, headers map[string]string) (*Response, error) {
	return h.request("GET", url, headers, nil)
}

func (h *HTTPRequest) POST(url string, headers map[string]string, rawBody map[string]any) (*Response, error) {
	if rawBody != nil {
		return h.request("POST", url, headers, &rawBody)
	}

	return h.request("POST", url, headers, nil)
}

func (h *HTTPRequest) PUT(url string, headers map[string]string, rawBody map[string]any) (*Response, error) {
	if rawBody != nil {
		return h.request("PUT", url, headers, &rawBody)
	}

	return h.request("PUT", url, headers, nil)
}

func (h *HTTPRequest) PATCH(url string, headers map[string]string, rawBody map[string]any) (*Response, error) {
	if rawBody != nil {
		return h.request("PATCH", url, headers, &rawBody)
	}

	return h.request("PATCH", url, headers, nil)
}

// change default timeout.
func (h *HTTPRequest) SetNewTimeout(timeout time.Duration) {
	h.timeout = timeout
}
