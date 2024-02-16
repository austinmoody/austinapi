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
	"time"
)

// TODO - too much converting going on just to get a sqid instead of the numeric database id, rethink this
//        has to be something simpler.

var (
	SleepRgxId   *regexp.Regexp
	SleepListRgx *regexp.Regexp

	InfoLog  *log.Logger
	ErrorLog *log.Logger
)

type SleepHandler struct{}

type Sleeps struct {
	Data          []Sleep `json:"data"`
	NextToken     string  `json:"next_token"`
	PreviousToken string  `json:"previous_token"`
}

type Sleep struct {
	ID               string
	Date             time.Time
	Rating           int64
	TotalSleep       int
	DeepSleep        int
	LightSleep       int
	RemSleep         int
	CreatedTimestamp time.Time
	UpdatedTimestamp time.Time
}

func ConvertSleepsRow(rows []austinapi_db.SleepsRow) []Sleep {
	var sleeps []Sleep

	for _, row := range rows {
		sleep := Sleep{}
		sleep.PopulateFromDbSleepRow(row)
		sleeps = append(sleeps, sleep)
	}

	return sleeps
}

func (s *Sleep) PopulateFromDbSleepRow(row austinapi_db.SleepsRow) {

	// TODO - handle potential errors from Encode
	s.ID, _ = IdHasher.Encode([]uint64{uint64(row.ID)})
	s.Date = row.Date
	s.Rating = row.Rating
	s.TotalSleep = row.TotalSleep
	s.DeepSleep = row.DeepSleep
	s.LightSleep = row.LightSleep
	s.RemSleep = row.LightSleep
	s.CreatedTimestamp = row.CreatedTimestamp
	s.UpdatedTimestamp = row.UpdatedTimestamp
}

func (s *Sleep) PopulateFromDbSleep(dbSleep austinapi_db.Sleep) {
	s.ID, _ = IdHasher.Encode([]uint64{uint64(dbSleep.ID)})
	s.Date = dbSleep.Date
	s.Rating = dbSleep.Rating
	s.TotalSleep = dbSleep.TotalSleep
	s.DeepSleep = dbSleep.DeepSleep
	s.LightSleep = dbSleep.LightSleep
	s.RemSleep = dbSleep.RemSleep
	s.CreatedTimestamp = dbSleep.CreatedTimestamp
	s.UpdatedTimestamp = dbSleep.UpdatedTimestamp
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

func (h *SleepHandler) GetSleep(w http.ResponseWriter, r *http.Request) {
	sleepIdMatches := SleepRgxId.FindStringSubmatch(r.URL.String())

	if len(sleepIdMatches) < 2 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, SleepRgxId.String())
		handleError(w, http.StatusInternalServerError, "Issue parsing specified id")
		return
	}

	sqid := sleepIdMatches[1]

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

	idFromSqid := GetIdFromToken(sqid)

	getSleepResult, err := apiDb.GetSleep(ctx, idFromSqid)

	if err != nil {
		ErrorLog.Printf("error retrieving sleep with id '%v': %v", idFromSqid, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(getSleepResult) != 1 {
		InfoLog.Printf("sleep with id '%v was not found in database", idFromSqid)
		handleError(w, http.StatusNotFound, fmt.Sprintf("Sleep not found with id %s", sqid))
		return
	}

	sleep := Sleep{}
	sleep.PopulateFromDbSleep(getSleepResult[0])
	jsonBytes, err := json.Marshal(sleep)
	if err != nil {
		ErrorLog.Printf("error marshaling JSON response: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
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
	sleeps.Data = ConvertSleepsRow(sleepsFromDb)
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
