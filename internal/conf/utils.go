package conf

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func GetRedisAddr() string {
	host := GetEnv("REDIS_HOST", "host.docker.internal")
	port := GetEnv("REDIS_PORT", "6379")
	addr := fmt.Sprintf("%s:%s", host, port)
	return addr
}

func LoadEnvFiles(args []string) error {
	var envFiles []string

	for i := 0; i < len(args); i++ {
		if args[i] == "--env-file" {
			if i+1 >= len(args) {
				// returning nil here, even though it's an error
				// because we want the caller to proceed anyway,
				// and produce the actual arg parsing error response
				return nil
			}

			envFiles = append(envFiles, args[i+1])
			i++
		}
	}

	if len(envFiles) == 0 {
		envFiles = []string{".env"}
	}

	// try to load all files in sequential order,
	// ignoring any that do not exist
	for _, file := range envFiles {
		err := godotenv.Load([]string{file}...)
		if err == nil {
			continue
		}

		var perr *os.PathError
		if errors.As(err, &perr) && errors.Is(perr, os.ErrNotExist) {
			// Ignoring file not found error
			continue
		}

		return fmt.Errorf("loading env file(s) %v: %v", envFiles, err)
	}

	return nil
}
