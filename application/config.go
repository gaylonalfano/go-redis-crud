package application

import (
	"os"
	"strconv"
)

type Config struct {
	RedisAddress string
	ServerPort   uint16
}

// Create a func to return an instance of our Config
// NOTE: viper and envconfig packages can do this as well
func LoadConfig() Config {
	// Create instance with defaults
	cfg := Config{
		RedisAddress: "localhost:6379",
		ServerPort:   3000,
	}

	// Import ENV variables using os package
	if redisAddr, exists := os.LookupEnv("REDIS_ADDR"); exists {
		// Overwrite defaults if it exists
		cfg.RedisAddress = redisAddr
	}

	if serverPort, exists := os.LookupEnv("SERVER_PORT"); exists {
		if port, err := strconv.ParseUint(serverPort, 10, 16); err == nil {
			cfg.ServerPort = uint16(port)
		}
	}

	return cfg
}
