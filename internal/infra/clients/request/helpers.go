package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
)

var errUnsupportedType = errors.New("type not supported")

func newErrorWrapper[T any](inputType T) error {
	return fmt.Errorf("%w: %T", errUnsupportedType, inputType)
}

func setFormData(rawBody *map[string]any) (io.Reader, string, error) {
	var (
		bodyBytes   bytes.Buffer
		body        io.Reader
		contentType string
		err         error
	)

	writer := multipart.NewWriter(&bodyBytes)
	defer func() {
		if err != nil {
			if err := writer.Close(); err != nil {
				log.Print(
					map[string]interface{}{
						"message": "erronr on close writer",
						"error":   err,
					},
				)
			}
		}
	}()

	var strVal string
	for key, val := range *rawBody {
		switch rawValue := val.(type) {
		case string:
			strVal = rawValue
		case int, int64, float64, bool:
			strVal = fmt.Sprintf("%v", rawValue)
		default:
			err = newErrorWrapper(rawValue)

			return nil, "", err
		}

		err = writer.WriteField(key, strVal)
		if err != nil {
			return nil, "", fmt.Errorf("error on write field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("error on close writer: %w", err)
	}

	body = &bodyBytes
	contentType = writer.FormDataContentType()

	return body, contentType, nil
}
