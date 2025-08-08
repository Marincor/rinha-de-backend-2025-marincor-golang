package constants

import (
	"net/http"
	"strconv"
)

const (
	HTTPStatusOK                     = http.StatusOK
	HTTPStatusCreated                = http.StatusCreated
	HTTPStatusAccepted               = http.StatusAccepted
	HTTPStatusNoContent              = http.StatusNoContent
	HTTPStatusBadRequest             = http.StatusBadRequest
	HTTPStatusUnauthorized           = http.StatusUnauthorized
	HTTPStatusNotAcceptable          = http.StatusNotAcceptable
	HTTPStatusTooManyRequests        = http.StatusTooManyRequests
	HTTPStatusInternalServerError    = http.StatusInternalServerError
	HTTPStatusForbidden              = http.StatusForbidden
	HTTPStatusUnprocessableEntity    = http.StatusUnprocessableEntity
	HTTPStatusNotFound               = http.StatusNotFound
	HTTPStatusConflict               = http.StatusConflict
	HTTPStatusGone                   = http.StatusGone
	HTTPStatusEntityTooLarge         = http.StatusRequestEntityTooLarge
	HTTPStatusUnsupportedMediaType   = http.StatusUnsupportedMediaType
	HTTPStatusStatusMovedPermanently = http.StatusMovedPermanently
)

var HTTPStatusesOk = []string{
	strconv.Itoa(HTTPStatusOK),
	strconv.Itoa(HTTPStatusCreated),
	strconv.Itoa(HTTPStatusAccepted),
	strconv.Itoa(HTTPStatusNoContent),
}
