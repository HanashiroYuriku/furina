package testingutils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"be-ayaka/config"
)

func GetDummyConfig() *config.Config {
	return &config.Config{
		Frontend: config.FrontendConfig{
			URL: "http://localhost:3000",
		},
	}
}

func StringPtr(s string) *string {
    return &s
}

func MakeJSONRequest(method, path string, body interface{}) *http.Request {
	var bodyReader io.Reader
	if body != nil {
		if b, ok := body.([]byte); ok {
			bodyReader = bytes.NewBuffer(b)
		} else {
			marshaled, _ := json.Marshal(body)
			bodyReader = bytes.NewBuffer(marshaled)
		}
	}

	req, _ := http.NewRequest(method, path, bodyReader)
	req.Host = "localhost"
	req.Header.Set("Content-Type", "application/json")

	return req
}