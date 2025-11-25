package database

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     int
	SSLMode  string
}

// LoadConfig loads DB config from environment variables or .env file
func LoadConfig() Config {
	// Cargamos el archivo .env, en gitignore, por lo que debemos tenerlo
	_ = godotenv.Load()

	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Printf("Warning: invalid DB_PORT value, defaulting to 5432")
		port = 5432
	}

	return Config{
		Host:     os.Getenv("DB_HOST"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		Port:     port,
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}
}
