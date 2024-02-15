package main

import (
	"github.com/joho/godotenv"
	"github.com/sqids/sqids-go"
	"log"
	"net/http"
	"os"
	"strconv"
)

// TODO Read these things from .evn: Port #, # of items to bring back in queries...

var (
	IdHasher   sqids.Sqids
	SqidLength string
)

func init() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	sqidKey := "SQID_LENGTH"
	SqidLength = os.Getenv(sqidKey)

	sqidLengthInteger, err := strconv.Atoi(SqidLength)
	if err != nil {
		log.Fatalf("Error converting %s to integer", sqidKey)
	}

	s, _ := sqids.New(sqids.Options{
		MinLength: uint8(sqidLengthInteger),
		Alphabet:  "usr4Z5gvSKhqpIt3BTAYVnwH8FQixC6G0cLNJ7fd9b1mlWEkOXz2RyjPoeUMDa",
	})

	IdHasher = *s
}

func main() {

	mux := http.NewServeMux()

	mux.Handle("/sleep", &SleepHandler{})
	mux.Handle("/sleep/", &SleepHandler{})

	http.ListenAndServe(":8080", mux)

}
