package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/austinmoody/austinapi_db/austinapi_db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

var (
	SleepRegex       = regexp.MustCompile(`^/sleep/*$`)
	SleepRegexWithID = regexp.MustCompile(`^/sleep/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$`)

	InfoLog  *log.Logger
	ErrorLog *log.Logger
)

type SleepHandler struct{}

type SleepList struct {
	Data          []austinapi_db.Sleep `json:"data"`
	NextToken     *string              `json:"next_token"`
	PreviousToken *string              `json:"previous_token"`
}

func init() {
	InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func (h *SleepHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet && SleepRegex.MatchString(r.URL.Path):
		h.ListSleep(w, r)
	case r.Method == http.MethodGet && SleepRegexWithID.MatchString(r.URL.Path):
		h.GetSleep(w, r)
	default:
		handleError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}

}

func handleError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(GenericMessage{Message: message})
}

func (h *SleepHandler) GetSleep(w http.ResponseWriter, r *http.Request) {
	sleepIdMatches := SleepRegexWithID.FindStringSubmatch(r.URL.Path)

	if len(sleepIdMatches) < 2 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, SleepRegexWithID.String())
		handleError(w, http.StatusInternalServerError, "Issue parsing specified id")
		return
	}

	connStr := GetDatabaseConnectionString()
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		ErrorLog.Printf("DB Connection error: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}
	defer conn.Close(ctx)

	apiDb := austinapi_db.New(conn)

	sleepUuid, err := uuid.Parse(sleepIdMatches[1])
	if err != nil {
		ErrorLog.Printf("error parsing UUID from http request: %v", err)
		handleError(w, http.StatusInternalServerError, fmt.Sprintf("Unable to parse specified id: %s", sleepIdMatches[1]))
		return
	}

	getSleepResult, err := apiDb.GetSleep(ctx, sleepUuid)

	if err != nil {
		ErrorLog.Printf("error retrieving sleep with id '%s': %v", sleepUuid.String(), err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(getSleepResult) != 1 {
		InfoLog.Printf("sleep with id '%s' was not found in database", sleepUuid.String())
		handleError(w, http.StatusNotFound, fmt.Sprintf("Sleep not found with id %s", sleepUuid.String()))
		return
	}

	result := getSleepResult[0]
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		ErrorLog.Printf("error marshaling JSON response: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *SleepHandler) ListSleep(w http.ResponseWriter, r *http.Request) {
	var sleepArray []austinapi_db.Sleep

	sleepArray = append(sleepArray, austinapi_db.Sleep{
		ID:               uuid.New(),
		Date:             time.Now(),
		Rating:           85,
		TotalSleep:       3534,
		DeepSleep:        456456,
		LightSleep:       234234,
		RemSleep:         24634,
		CreatedTimestamp: time.Now(),
		UpdatedTimestamp: time.Now(),
	})

	sleepList := SleepList{
		Data:          sleepArray,
		NextToken:     nil,
		PreviousToken: nil,
	}

	jsonBytes, err := json.Marshal(sleepList)
	if err != nil {
		log.Fatalf("%v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)

}
