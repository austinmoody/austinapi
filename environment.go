package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

func GetString(key string) string {
	err := godotenv.Load()
	if err != nil {
		log.Printf(".env not found, using environment")
	}

	return os.Getenv(key)
}

func GetInt(key string) int {
	err := godotenv.Load()
	if err != nil {
		log.Printf(".env not found, using environment")
	}

	returnInt, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		log.Fatalf("error converting environment value %s to integer", key)
	}

	return returnInt
}

func GetInt32(key string) int32 {
	returnInt := GetInt(key)
	return int32(returnInt)
}

func GetUint8(key string) uint8 {
	returnInt := GetInt(key)
	return uint8(returnInt)
}
