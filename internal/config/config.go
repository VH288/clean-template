package config

import (
	"log"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName   string
	AppSecret string
	Port      string
	GRPCPort  string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	ExternalAPIURL  string
	ExternalGRPCAddr string
}

func Load() Config {
	env, err := godotenv.Read(".env")
	if err != nil {
		log.Fatal("failed to read env file: ", err)
	}

	return Config{
		AppName:        getEnv(env, "APP_NAME", "clean-template"),
		AppSecret:      getEnv(env, "APP_SECRET", ""),
		Port:           getEnv(env, "PORT", "8080"),
		GRPCPort:       getEnv(env, "GRPC_PORT", "7000"),
		DBHost:         getEnv(env, "DB_HOST", "127.0.0.1"),
		DBPort:         getEnv(env, "DB_PORT", "3306"),
		DBName:         getEnv(env, "DB_NAME", ""),
		DBUser:         getEnv(env, "DB_USER", ""),
		DBPassword:     getEnv(env, "DB_PASSWORD", ""),
		ExternalAPIURL:   getEnv(env, "EXTERNAL_API_URL", "https://jsonplaceholder.typicode.com"),
		ExternalGRPCAddr: getEnv(env, "EXTERNAL_GRPC_ADDR", "localhost:7001"),
	}
}

func getEnv(env map[string]string, key, fallback string) string {
	if val := env[key]; val != "" {
		return val
	}
	return fallback
}
