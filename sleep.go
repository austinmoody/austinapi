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

// TODO - create requestId to tie things together in the logs

var (
	SleepRgxId   *regexp.Regexp
	SleepListRgx *regexp.Regexp
	SleepRgxDate *regexp.Regexp

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

type SleepsResult []austinapi_db.SleepsRow
type SleepsItem austinapi_db.SleepsRow

func (si SleepsItem) ToSleep() Sleep {
	var sleep Sleep
	// TODO - handle potential errors from Encode
	sleep.ID, _ = IdHasher.Encode([]uint64{uint64(si.ID)})
	sleep.Date = si.Date
	sleep.Rating = si.Rating
	sleep.TotalSleep = si.TotalSleep
	sleep.DeepSleep = si.DeepSleep
	sleep.LightSleep = si.LightSleep
	sleep.RemSleep = si.LightSleep
	sleep.CreatedTimestamp = si.CreatedTimestamp
	sleep.UpdatedTimestamp = si.UpdatedTimestamp

	return sleep
}

func (sr SleepsResult) ToSleeps() Sleeps {
	var sleeps Sleeps

	var data []Sleep

	for _, row := range sr {
		data = append(data, SleepsItem(row).ToSleep())
	}

	sleeps.Data = data

	sleeps.NextToken = sr.GetNextToken()
	sleeps.PreviousToken = sr.GetPreviousToken()

	return sleeps
}

type SleepResult austinapi_db.Sleep

func (s SleepResult) ToSleep() Sleep {
	var sleep Sleep

	sleep.ID, _ = IdHasher.Encode([]uint64{uint64(s.ID)})
	sleep.Date = s.Date
	sleep.Rating = s.Rating
	sleep.TotalSleep = s.TotalSleep
	sleep.DeepSleep = s.DeepSleep
	sleep.LightSleep = s.LightSleep
	sleep.RemSleep = s.RemSleep
	sleep.CreatedTimestamp = s.CreatedTimestamp
	sleep.UpdatedTimestamp = s.UpdatedTimestamp

	return sleep
}

func (sr SleepsResult) GetNextToken() string {
	var nextToken string

	nextTokenInt := sr[len(sr)-1].NextID
	if nextTokenInt < 1 {
		nextToken = ""
	} else {
		id, _ := IdHasher.Encode([]uint64{uint64(nextTokenInt)})
		nextToken = id
	}

	return nextToken
}

func (sr SleepsResult) GetPreviousToken() string {
	var previousToken string
	previousTokenInt := sr[0].PreviousID
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

	SleepRgxId = regexp.MustCompile(fmt.Sprintf(`^/sleep/id/([0-9a-zA-Z]{%s})$`, SqidLength))
	SleepListRgx = regexp.MustCompile(fmt.Sprintf(`^/sleep/list(?:\?(next_token|previous_token)=([0-9a-zA-Z]{%s}))?$`, SqidLength))
	SleepRgxDate = regexp.MustCompile(`^/sleep/date/([0-9]{4}-[0-9]{2}-[0-9]{2})$`)
}

func (h *SleepHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet && SleepListRgx.MatchString(r.URL.String()):
		h.ListSleep(w, r)
	case r.Method == http.MethodGet && SleepRgxId.MatchString(r.URL.String()):
		h.GetSleep(w, r)
	case r.Method == http.MethodGet && SleepRgxDate.MatchString(r.URL.String()):
		h.GetSleepByDate(w, r)
	default:
		handleError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}

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

// @Summary Get sleep information by ID
// @Security ApiKeyAuth
// @Description Retrieves sleep information with specified ID
// @Tags sleep
// @Accept json
// @Produce json
// @Param id path string true "Sleep ID"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} Sleep
// @Failure 500 {object} GenericMessage
// @Failure 404 {object} GenericMessage
// @Failure 401
// @Router /sleep/id/{id} [get]
func (h *SleepHandler) GetSleep(w http.ResponseWriter, r *http.Request) {
	sleepIdMatches := SleepRgxId.FindStringSubmatch(r.URL.String())

	if len(sleepIdMatches) < 2 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, SleepRgxId.String())
		handleError(w, http.StatusInternalServerError, "Issue parsing specified id")
		return
	}

	InfoLog.Printf("URL token match '%s'\n", sleepIdMatches[1])

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

	jsonBytes, err := json.Marshal(SleepResult(getSleepResult[0]).ToSleep())
	if err != nil {
		ErrorLog.Printf("error marshaling JSON response: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonBytes)
	if err != nil {
		ErrorLog.Printf("error writing http response: %v\n", err)
	}

}

// @Summary Get sleep information by date
// @Security ApiKeyAuth
// @Description Retrieves sleep information with specified date
// @Tags sleep
// @Accept json
// @Produce json
// @Param date path string true "Date"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} Sleep
// @Failure 500 {object} GenericMessage
// @Failure 404 {object} GenericMessage
// @Failure 401
// @Router /sleep/date/{date} [get]
func (h *SleepHandler) GetSleepByDate(w http.ResponseWriter, r *http.Request) {
	sleepDateMatches := SleepRgxDate.FindStringSubmatch(r.URL.String())

	if len(sleepDateMatches) < 2 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, SleepRgxId.String())
		handleError(w, http.StatusInternalServerError, "Issue parsing specified date")
		return
	}

	InfoLog.Printf("URL token match '%s'\n", sleepDateMatches[1])

	sleepDateString := sleepDateMatches[1]
	sleepDate, err := time.Parse("2006-01-02", sleepDateString)
	if err != nil {
		ErrorLog.Printf("Unable to parse '%s' to time.Time object: %v", sleepDateString, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
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

	getSleepResult, err := apiDb.GetSleepByDate(ctx, sleepDate)

	if err != nil {
		ErrorLog.Printf("error retrieving sleep with date '%v': %v", sleepDateString, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(getSleepResult) != 1 {
		InfoLog.Printf("sleep with date '%s' was not found in database", sleepDateString)
		handleError(w, http.StatusNotFound, fmt.Sprintf("Sleep not found with date %s", sleepDateString))
		return
	}

	jsonBytes, err := json.Marshal(SleepResult(getSleepResult[0]).ToSleep())
	if err != nil {
		ErrorLog.Printf("error marshaling JSON response: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonBytes)
	if err != nil {
		ErrorLog.Printf("error writing http response: %v\n", err)
	}

}

// @Summary Get list of sleep information
// @Security ApiKeyAuth
// @Description Retrieves list of sleep information in descending order by date
// @Description Specifying no query parameters pulls list starting with latest
// @Description Caller can then specify a next_token or previous_token returned from
// @Description calls to go forward and back in the list of items.  Only next_token OR
// @Description previous_token should be specified.
// @Tags sleep
// @Produce json
// @Param next_token query string false "next list search by next_token" Format(string)
// @Param previous_token query string false "previous list search by previous_token" Format(string)
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} Sleeps
// @Failure 500 {object} GenericMessage
// @Failure 401
// @Router /sleep/list [get]
func (h *SleepHandler) ListSleep(w http.ResponseWriter, r *http.Request) {
	urlMatches := SleepListRgx.FindStringSubmatch(r.URL.String())

	if len(urlMatches) != 3 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, SleepListRgx.String())
		handleError(w, http.StatusInternalServerError, "Issue parsing URL")
		return
	}

	queryType := urlMatches[1]
	queryToken := urlMatches[2]

	InfoLog.Printf("URL directive match '%s'\n", queryType)
	InfoLog.Printf("URL token match '%s'\n", queryToken)

	connStr := GetDatabaseConnectionString()
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		ErrorLog.Printf("database connection error: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}
	defer conn.Close(ctx)

	apiDb := austinapi_db.New(conn)

	params := austinapi_db.SleepsParams{
		QueryType: "",
		InputID:   0,
		RowLimit:  ListRowLimit,
	}

	switch queryType {
	case "next_token":
		params.QueryType = "NEXT"
		params.InputID = GetIdFromToken(queryToken)
	case "previous_token":
		params.QueryType = "PREVIOUS"
		params.InputID = GetIdFromToken(queryToken)
	}

	sleepsFromDb, err := apiDb.Sleeps(ctx, params)
	if err != nil {
		ErrorLog.Printf("error getting list of sleep: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(sleepsFromDb) < 1 {
		ErrorLog.Printf("no results from database with '%' token '%s'", queryType, queryToken)
		handleError(w, http.StatusNotFound, "no results found")
		return
	}

	jsonBytes, err := json.Marshal(SleepsResult(sleepsFromDb).ToSleeps())
	if err != nil {
		ErrorLog.Printf("error marshaling JSON response: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonBytes)
	if err != nil {
		ErrorLog.Printf("error writing http response: %v", err)
	}
}
