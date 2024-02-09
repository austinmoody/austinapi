package main

import (
	"context"
	"encoding/json"
	"github.com/austinmoody/austinapi_db/austinapi_db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"regexp"
	"time"
)

var (
	SleepRegex       = regexp.MustCompile(`^/sleep/*$`)
	SleepRegexWithID = regexp.MustCompile(`^/sleep/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$`)
)

type SleepHandler struct{}

type SleepList struct {
	Data          []austinapi_db.Sleep `json:"data"`
	NextToken     *string              `json:"next_token"`
	PreviousToken *string              `json:"previous_token"`
}

func (h *SleepHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && SleepRegex.MatchString(r.URL.Path):
		h.ListSleep(w, r)
	case r.Method == http.MethodGet && SleepRegexWithID.MatchString(r.URL.Path):
		h.GetSleep(w, r)
	default:
		return
	}
}

func (h *SleepHandler) GetSleep(w http.ResponseWriter, r *http.Request) {

	sleepIdMatches := SleepRegexWithID.FindStringSubmatch(r.URL.Path)

	if len(sleepIdMatches) < 2 {
		// InternalServerErrorHandler(w, r) // TODO implement
		return
	}

	connStr := GetDatabaseConnectionString()
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("DB Connection error: %v", err)
	}
	defer conn.Close(ctx)

	apiDb := austinapi_db.New(conn)

	sleepUuid, err := uuid.Parse(sleepIdMatches[1])
	if err != nil {
		log.Fatalf("Error getting UUID: %v", sleepUuid)
	}

	mySleep, err := apiDb.GetSleep(ctx, sleepUuid)
	// TODO not fatal error
	if err != nil {
		log.Fatalf("Error getting Sleep with ID %v", mySleep)
	}

	jsonBytes, err := json.Marshal(mySleep)
	if err != nil {
		log.Fatalf("%v", err)
	}

	w.Header().Set("Content-Type", "application/json")
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
