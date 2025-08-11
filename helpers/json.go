package helpers

import (
	json "github.com/bytedance/sonic"
)

type JSONRawMessage json.NoCopyRawMessage

var (
	Marshal    = json.Marshal
	Unmarshal  = json.Unmarshal
	NewDecoder = json.ConfigDefault.NewDecoder
	NewEncoder = json.ConfigDefault.NewEncoder
)
