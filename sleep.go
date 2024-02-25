package main

import (
	"encoding/json"
	"fmt"
	"github.com/austinmoody/austinapi_db/austinapi_db"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// TODO - create requestId to tie things together in the logs

var (
	SleepRgxId   *regexp.Regexp
	SleepListRgx *regexp.Regexp
	SleepRgxDate *regexp.Regexp
)

type SleepHandler struct{}

type Sleeps struct {
	Data      []austinapi_db.Sleep `json:"data"`
	NextToken int32                `json:"next_token"`
}

func init() {
	SleepRgxId = regexp.MustCompile(`^/sleep/id/([0-9]+)$`)
	SleepListRgx = regexp.MustCompile(`^/sleep/list(?:\?(next_token)=([0-9]+))?$`)
	SleepRgxDate = regexp.MustCompile(`^/sleep/date/([0-9]{4}-[0-9]{2}-[0-9]{2})$`)
}

func (h *SleepHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet && SleepListRgx.MatchString(r.URL.String()):
		h.listSleep(w, r)
	case r.Method == http.MethodGet && SleepRgxId.MatchString(r.URL.String()):
		h.getSleep(w, r)
	case r.Method == http.MethodGet && SleepRgxDate.MatchString(r.URL.String()):
		h.getSleepByDate(w, r)
	default:
		handleError(w, http.StatusMethodNotAllowed, "Method not allowed")
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
// @Success 200 {object} austinapi_db.Sleep
// @Failure 500 {object} GenericMessage
// @Failure 404 {object} GenericMessage
// @Failure 401
// @Router /sleep/id/{id} [get]
func (h *SleepHandler) getSleep(w http.ResponseWriter, r *http.Request) {

	id, err := getIdFromUrl(SleepRgxId, r.URL)

	if err != nil {
		ErrorLog.Println(err)
		handleError(w, http.StatusInternalServerError, "Issue parsing id from URL")
		return
	}

	InfoLog.Printf("URL id match '%d'\n", id)

	result, err := ApiDatabase.GetSleep(DatabaseContext, id)

	if err != nil {
		ErrorLog.Printf("error retrieving sleep with id '%d': %v", id, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(result) != 1 {
		InfoLog.Printf("sleep with id '%d' was not found in database", id)
		handleError(w, http.StatusNotFound, fmt.Sprintf("Sleep not found with id %d", id))
		return
	}

	jsonBytes, err := json.Marshal(result[0])
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
// @Success 200 {object} austinapi_db.Sleep
// @Failure 500 {object} GenericMessage
// @Failure 404 {object} GenericMessage
// @Failure 401
// @Router /sleep/date/{date} [get]
func (h *SleepHandler) getSleepByDate(w http.ResponseWriter, r *http.Request) {
	sleepDateMatches := SleepRgxDate.FindStringSubmatch(r.URL.String())

	if len(sleepDateMatches) < 2 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, SleepRgxDate.String())
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

	result, err := ApiDatabase.GetSleepByDate(DatabaseContext, sleepDate)

	if err != nil {
		ErrorLog.Printf("error retrieving sleep with date '%v': %v", sleepDateString, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(result) != 1 {
		InfoLog.Printf("sleep with date '%s' was not found in database", sleepDateString)
		handleError(w, http.StatusNotFound, fmt.Sprintf("Sleep not found with date %s", sleepDateString))
		return
	}

	jsonBytes, err := json.Marshal(result[0])
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
// @Description Caller can then specify a next_token from previous calls to go
// @Description forward in the list of items.
// @Tags sleep
// @Produce json
// @Param next_token query string false "next list search by next_token" Format(string)
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} Sleeps
// @Failure 500 {object} GenericMessage
// @Failure 401
// @Router /sleep/list [get]
func (h *SleepHandler) listSleep(w http.ResponseWriter, r *http.Request) {
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

	params := austinapi_db.GetSleepsParams{
		RowOffset: 0,
		RowLimit:  ListRowLimit,
	}

	if queryType == "next_token" {
		rowOffset, err := strconv.ParseInt(queryToken, 10, 32)
		if err != nil {
			ErrorLog.Printf("error parsing specified query token '%v': %v", queryToken, err)
			handleError(w, http.StatusBadRequest, "Invalid query token")
			return
		}

		params.RowOffset = int32(rowOffset)
	}

	results, err := ApiDatabase.GetSleeps(DatabaseContext, params)
	if err != nil {
		ErrorLog.Printf("error getting list of sleep: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(results) < 1 {
		ErrorLog.Printf("no sleep results from database with '%s' token '%s'", queryType, queryToken)
		handleError(w, http.StatusNotFound, "no results found")
		return
	}

	sleeps := Sleeps{
		Data:      results,
		NextToken: params.RowLimit + params.RowOffset,
	}

	jsonBytes, err := json.Marshal(sleeps)
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
