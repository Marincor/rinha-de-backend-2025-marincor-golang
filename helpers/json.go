package helpers

import (
	"sync"

	json "github.com/bytedance/sonic"
)

type JSONRawMessage json.NoCopyRawMessage

// Pool de buffers para reutilizar memory allocations.
var (
	bufferPool = sync.Pool{
		New: func() interface{} {
			// Buffer inicial de 1KB, vai crescer conforme necessário
			return make([]byte, 0, 1024)
		},
	}

	config = json.Config{
		NoQuoteTextMarshaler:    true,
		NoNullSliceOrMap:        true,
		ValidateString:          false,
		NoValidateJSONMarshaler: true,
	}.Froze()

	NewEncoder = config.NewEncoder

	NewDecoder = config.NewDecoder
)

func Unmarshal(data []byte, v interface{}) error {
	return config.Unmarshal(data, v)
}

func Marshal(value interface{}) ([]byte, error) {
	buf := bufferPool.Get().([]byte)

	buf = buf[:0] // Reset length.

	defer bufferPool.Put(buf)

	result, err := config.Marshal(value)
	if err != nil {
		return nil, err
	}

	// Copy to avoid reference to buffer pool
	output := make([]byte, len(result))
	copy(output, result)

	return output, nil
}
