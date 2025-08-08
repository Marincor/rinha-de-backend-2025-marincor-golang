package helpers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func CreateResponse(context *fiber.Ctx, payload interface{}, status ...int) error {
	returnStatus := http.StatusOK
	if len(status) > 0 {
		returnStatus = status[0]
	}

	return context.Status(returnStatus).JSON(payload)
}

// AllParams Params is used to get all route parameters.
func AllParams(context *fiber.Ctx) map[string]string {
	route := context.Route()

	if len(route.Params) == 0 {
		return nil
	}

	return context.AllParams()
}

type SuccessListResponse struct {
	Data  interface{} `json:"data"`
	Count int         `json:"count"`
}

type ErrorResponse struct {
	Message     string `json:"error"`
	Description string `json:"description,omitempty"`
	EscapeURL   string `json:"escape_url,omitzero"`
	StatusCode  int    `json:"status_code,omitempty"`
}
