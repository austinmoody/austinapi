package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sqids/sqids-go"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"
)

var (
	IdHasher      sqids.Sqids
	SqidLength    string
	ListeningPort string
	ListRowLimit  int32
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
)

func init() {

	err := godotenv.Load()
	if err != nil {
		log.Printf(".env not found, using environment")
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

	InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

}

func main() {

	mux := http.NewServeMux()

	// Serve Swagger UI files
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

	mux.Handle("/sleep", authenticator(&SleepHandler{}))
	mux.Handle("/sleep/", authenticator(&SleepHandler{}))

	mux.Handle("/readyscore", authenticator(&ReadyScoreHandler{}))
	mux.Handle("/readyscore/", authenticator(&ReadyScoreHandler{}))

	http.ListenAndServe(ListeningPort, mux)

}

func GetIdFromToken(token string) int64 {
	nextTokenSlice := IdHasher.Decode(token)
	return int64(nextTokenSlice[0])
}
