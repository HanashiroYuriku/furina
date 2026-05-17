package testingutils

import (
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
