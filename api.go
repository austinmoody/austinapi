package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sqids/sqids-go"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
)

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

	// Serve Swagger UI files
	//mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("swagger-ui"))))
	//mux.Handle("/swagger/", httpSwagger.WrapHandler)
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.json"), // URL to the Swagger JSON file
	))

	// Serve Swagger JSON at /swagger.json
	mux.HandleFunc("/swagger/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.json")
	})

	mux.HandleFunc("/swagger/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger.yaml")
	})

	//mux.Handle("/sleep", &SleepHandler{})
	//mux.Handle("/sleep/", &SleepHandler{})

	mux.Handle("/sleep", authenticator(&SleepHandler{}))
	mux.Handle("/sleep/", authenticator(&SleepHandler{}))

	http.ListenAndServe(ListeningPort, mux)

}
