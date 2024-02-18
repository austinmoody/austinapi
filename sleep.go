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

// @Summary Get sleep information
// @Description Retrieves sleep information
// @Tags sleep
// @Accept json
// @Produce json
// @Param id path string true "Sleep ID"
// @Success 200 {object} Sleep
// @Router /sleep/{id} [get]
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
		ErrorLog.Printf("error getting list of sleep: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
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
