package common

import (
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	godotenv.Load()
}

func GetEnv(key string) string{
	value, _ := os.LookupEnv(key)
	return value
}