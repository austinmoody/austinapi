package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"os"
)

var (
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

func handleError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)

	jsonBytes, err := json.Marshal(GenericMessage{Message: message})
	if err != nil {
		ErrorLog.Printf("error marshaling JSON error response: %v", err)
		return
	}
	_, err = w.Write(jsonBytes)
	if err != nil {
		ErrorLog.Printf("error writing error response: %v", err)
	}
}
