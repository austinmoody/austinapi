package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/austinmoody/austinapi_db/austinapi_db"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
)

var (
	ListeningPort      string
	ListRowLimit       int32
	InfoLog            *log.Logger
	ErrorLog           *log.Logger
	DatabaseContext    context.Context
	DatabaseConnection *pgx.Conn
	ApiDatabase        *austinapi_db.Queries
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

	connStr := getDatabaseConnectionString()
	DatabaseContext = context.Background()

	DatabaseConnection, err = pgx.Connect(DatabaseContext, connStr)
	if err != nil {
		log.Fatalf("DB Connection error: %v", err)
	}

	ApiDatabase = austinapi_db.New(DatabaseConnection)
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

	defer DatabaseConnection.Close(DatabaseContext)

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

func getIdFromUrl(regex *regexp.Regexp, url *url.URL) (int64, error) {
	matches := regex.FindStringSubmatch(url.String())

	if len(matches) < 2 {
		return -1, fmt.Errorf("no ID found in URL path '%s'", url.String())
	}

	id, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return -1, fmt.Errorf("unable to convert ID '%s' to integer: %v", matches[1], err)
	}

	return id, nil
}

func getDatabaseConnectionString() string {
	err := godotenv.Load()
	if err != nil {
		log.Printf(".env not found, using environment")
	}

	databaseHost := os.Getenv("DATABASE_HOST")
	databasePort := os.Getenv("DATABASE_PORT")
	databaseUser := os.Getenv("DATABASE_USER")
	databasePassword := os.Getenv("DATABASE_PASSWORD")
	databaseName := os.Getenv("DATABASE_NAME")
	sslMode := os.Getenv("DATABASE_SSLMODE")

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		databaseHost,
		databasePort,
		databaseUser,
		databasePassword,
		databaseName,
		sslMode,
	)
}
