package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sqids/sqids-go"
	"log"
	"net/http"
)

// TODO Read these things from .evn: Port #, # of items to bring back in queries...

var (
	IdHasher      sqids.Sqids
	SqidLength    string
	ListeningPort string
	ListRowLimit  int32
)

func init() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	sqidLength := GetUint8("SQID_LENGTH")
	s, _ := sqids.New(sqids.Options{
		MinLength: sqidLength,
		Alphabet:  "usr4Z5gvSKhqpIt3BTAYVnwH8FQixC6G0cLNJ7fd9b1mlWEkOXz2RyjPoeUMDa",
	})
	SqidLength = GetString("SQID_LENGTH")

	IdHasher = *s

	ListeningPort = fmt.Sprintf(":%s", GetString("LISTENING_PORT"))

	ListRowLimit = GetInt32("LIST_ROW_LIMIT")

}

func main() {

	mux := http.NewServeMux()

	mux.Handle("/sleep", &SleepHandler{})
	mux.Handle("/sleep/", &SleepHandler{})

	http.ListenAndServe(ListeningPort, mux)

}
