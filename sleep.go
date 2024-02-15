package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/austinmoody/austinapi_db/austinapi_db"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"os"
	"regexp"
)

var (
	SleepRgxId   *regexp.Regexp
	SleepListRgx *regexp.Regexp

	InfoLog  *log.Logger
	ErrorLog *log.Logger
)

type SleepHandler struct{}

// TODO custom Marshal to sqid id's
type Sleeps struct {
	Data          []austinapi_db.SleepsRow `json:"data"`
	NextToken     string                   `json:"next_token"`
	PreviousToken string                   `json:"previous_token"`
}

func GetNextToken(sleeps []austinapi_db.SleepsRow) string {
	var nextToken string

	nextTokenInt := sleeps[len(sleeps)-1].NextID
	if nextTokenInt < 1 {
		nextToken = ""
	} else {
		id, _ := IdHasher.Encode([]uint64{uint64(nextTokenInt)})
		nextToken = id
	}

	return nextToken
}

func GetPreviousToken(sleeps []austinapi_db.SleepsRow) string {
	var previousToken string
	previousTokenInt := sleeps[0].PreviousID
	if previousTokenInt < 1 {
		previousToken = ""
	} else {
		id, _ := IdHasher.Encode([]uint64{uint64(previousTokenInt)})
		previousToken = id
	}

	return previousToken
}

func GetIdFromToken(token string) int64 {
	nextTokenSlice := IdHasher.Decode(token)
	return int64(nextTokenSlice[0])
}

func init() {
	InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	SleepRgxId = regexp.MustCompile(fmt.Sprintf(`^/sleep/([0-9a-zA-Z]{%s})$`, SqidLength))
	SleepListRgx = regexp.MustCompile(fmt.Sprintf(`^/sleep(?:\?(next_token|previous_token)=([0-9a-zA-Z]{%s}))?$`, SqidLength))

}

func (h *SleepHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet && SleepListRgx.MatchString(r.URL.String()):
		h.ListSleep(w, r)
	case r.Method == http.MethodGet && SleepRgxId.MatchString(r.URL.String()):
		h.GetSleep(w, r)
	default:
		handleError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}

}

func handleError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(GenericMessage{Message: message})
}

// TODO fix GetSleep to use sqlid
func (h *SleepHandler) GetSleep(w http.ResponseWriter, r *http.Request) {
	//sleepIdMatches := SleepRgxId.FindStringSubmatch(r.URL.String())
	//
	//if len(sleepIdMatches) < 2 {
	//	ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, SleepRgxId.String())
	//	handleError(w, http.StatusInternalServerError, "Issue parsing specified id")
	//	return
	//}
	//
	//connStr := GetDatabaseConnectionString()
	//ctx := context.Background()
	//
	//conn, err := pgx.Connect(ctx, connStr)
	//if err != nil {
	//	ErrorLog.Printf("DB Connection error: %v", err)
	//	handleError(w, http.StatusInternalServerError, "Internal Error")
	//	return
	//}
	//defer conn.Close(ctx)
	//
	//apiDb := austinapi_db.New(conn)
	//
	//sleepUuid, err := uuid.Parse(sleepIdMatches[1])
	//if err != nil {
	//	ErrorLog.Printf("error parsing UUID from http request: %v", err)
	//	handleError(w, http.StatusInternalServerError, fmt.Sprintf("Unable to parse specified id: %s", sleepIdMatches[1]))
	//	return
	//}
	//
	//getSleepResult, err := apiDb.GetSleep(ctx, sleepUuid)
	//
	//if err != nil {
	//	ErrorLog.Printf("error retrieving sleep with id '%s': %v", sleepUuid.String(), err)
	//	handleError(w, http.StatusInternalServerError, "Internal Error")
	//	return
	//}
	//
	//if len(getSleepResult) != 1 {
	//	InfoLog.Printf("sleep with id '%s' was not found in database", sleepUuid.String())
	//	handleError(w, http.StatusNotFound, fmt.Sprintf("Sleep not found with id %s", sleepUuid.String()))
	//	return
	//}
	//
	//result := getSleepResult[0]
	//jsonBytes, err := json.Marshal(result)
	//if err != nil {
	//	ErrorLog.Printf("error marshaling JSON response: %v", err)
	//	handleError(w, http.StatusInternalServerError, "Internal Error")
	//	return
	//}
	//
	//w.WriteHeader(http.StatusOK)
	//w.Write(jsonBytes)
}

func (h *SleepHandler) ListSleep(w http.ResponseWriter, r *http.Request) {
	urlMatches := SleepListRgx.FindStringSubmatch(r.URL.String())

	if len(urlMatches) != 3 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, SleepListRgx.String())
		handleError(w, http.StatusInternalServerError, "Issue parsing URL")
		return
	}

	InfoLog.Printf("URL directive match '%s'\n", urlMatches[1])
	InfoLog.Printf("URL token match '%s'\n", urlMatches[2])

	connStr := GetDatabaseConnectionString()
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("DB Connection error: %v", err)
	}
	defer conn.Close(ctx)

	apiDb := austinapi_db.New(conn)

	params := austinapi_db.SleepsParams{
		QueryType: "",
		InputID:   0,
		RowLimit:  ListRowLimit,
	}

	switch urlMatches[1] {
	case "next_token":
		params.QueryType = "NEXT"
		params.InputID = GetIdFromToken(urlMatches[2])
	case "previous_token":
		params.QueryType = "PREVIOUS"
		params.InputID = GetIdFromToken(urlMatches[2])
	}

	sleepsFromDb, err := apiDb.Sleeps(ctx, params)
	if err != nil {
		log.Fatalf("Error getting list of sleep %v\n", err)
	}

	sleeps := Sleeps{}
	sleeps.Data = sleepsFromDb
	sleeps.NextToken = GetNextToken(sleepsFromDb)
	sleeps.PreviousToken = GetPreviousToken(sleepsFromDb)

	jsonBytes, err := json.Marshal(sleeps)
	if err != nil {
		log.Fatalf("error marshaling JSON response: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)

}
