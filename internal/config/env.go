package config

import (
	"os"
	"strconv"
)

func getEnvString(env string, val *string) {
	value := os.Getenv(env)
	if len(value) == 0 {
		return
	}

	*val = value
}

func getEnvUint(env string, val *uint) {
	s := os.Getenv(env)
	if len(s) == 0 {
		return
	}

	value, err := strconv.ParseUint(s, 10, 0)
	if err == nil {
		*val = uint(value)
	}
}

func (c *RootConfig) FromEnvironment() {
	getEnvString("LW_EXPORTER_ADDRESS", &c.Address)
	getEnvString("LW_EXPORTER_API_KEY", &c.ApiKey)
	getEnvUint("LW_EXPORTER_PORT", &c.Port)
}
