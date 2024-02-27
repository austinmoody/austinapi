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

var (
	HeartRateRgxId   *regexp.Regexp
	HeartRateListRgx *regexp.Regexp
	HeartRateRgxDate *regexp.Regexp
)

type HeartRateHandler struct{}

type HeartRates struct {
	Data      []austinapi_db.Heartrate `json:"data"`
	NextToken int32                    `json:"next_token"`
}

func init() {
	HeartRateRgxId = regexp.MustCompile(`^/heartrate/id/([0-9]+)$`)
	HeartRateListRgx = regexp.MustCompile(`^/heartrate/list(?:\?(next_token)=([0-9]+))?$`)
	HeartRateRgxDate = regexp.MustCompile(`^/heartrate/date/([0-9]{4}-[0-9]{2}-[0-9]{2})$`)
}

func (h *HeartRateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet && HeartRateListRgx.MatchString(r.URL.String()):
		h.listHeartRate(w, r)
	case r.Method == http.MethodGet && HeartRateRgxId.MatchString(r.URL.String()):
		h.getHeartRate(w, r)
	case r.Method == http.MethodGet && HeartRateRgxDate.MatchString(r.URL.String()):
		h.getHeartRateByDate(w, r)
	default:
		handleError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// @Summary Get heart rate information by ID
// @Security ApiKeyAuth
// @Description Retrieves heart rate information with specified ID
// @Tags heartrate
// @Accept json
// @Produce json
// @Param id path string true "Heart Rate ID"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} austinapi_db.Heartrate
// @Failure 500 {object} GenericMessage
// @Failure 404 {object} GenericMessage
// @Failure 401
// @Router /heartrate/id/{id} [get]
func (h *HeartRateHandler) getHeartRate(w http.ResponseWriter, r *http.Request) {
	id, err := getIdFromUrl(HeartRateRgxId, r.URL)

	if err != nil {
		ErrorLog.Println(err)
		handleError(w, http.StatusInternalServerError, "Issue parsing id from URL")
		return
	}

	InfoLog.Printf("URL id match '%d'\n", id)

	result, err := ApiDatabase.GetHeartRate(DatabaseContext, id)

	if err != nil {
		ErrorLog.Printf("error retrieving heart rate with id '%d': %v", id, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(result) != 1 {
		InfoLog.Printf("heart rate with id '%d' was not found in database", id)
		handleError(w, http.StatusNotFound, fmt.Sprintf("Heart Rate not found with id %d", id))
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

// @Summary Get heart rate information by date
// @Security ApiKeyAuth
// @Description Retrieves heart rate information with specified date
// @Tags heartrate
// @Accept json
// @Produce json
// @Param date path string true "Date"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} austinapi_db.Heartrate
// @Failure 500 {object} GenericMessage
// @Failure 404 {object} GenericMessage
// @Failure 401
// @Router /heartrate/date/{date} [get]
func (h *HeartRateHandler) getHeartRateByDate(w http.ResponseWriter, r *http.Request) {

	dateMatches := HeartRateRgxDate.FindStringSubmatch(r.URL.String())

	if len(dateMatches) < 2 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, HeartRateRgxDate.String())
		handleError(w, http.StatusInternalServerError, "Issue parsing specified date")
		return
	}

	InfoLog.Printf("URL token match '%s'\n", dateMatches[1])

	dateString := dateMatches[1]
	date, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		ErrorLog.Printf("Unable to parse '%s' to time.Time object: %v", dateString, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	result, err := ApiDatabase.GetHeartRateByDate(DatabaseContext, date)

	if err != nil {
		ErrorLog.Printf("error retrieving heart rate with date '%v': %v", dateString, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(result) != 1 {
		InfoLog.Printf("heart rate with date '%s' was not found in database", dateString)
		handleError(w, http.StatusNotFound, fmt.Sprintf("Heart Rate not found with date %s", dateString))
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

// @Summary Get list of heart rate information
// @Security ApiKeyAuth
// @Description Retrieves list of heart rate information in descending order by date
// @Description Specifying no query parameters pulls list starting with latest
// @Description Caller can then specify a next_token from previous calls to go
// @Description forward in the list of items.
// @Tags heartrate
// @Produce json
// @Param next_token query string false "next list search by next_token" Format(string)
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} HeartRates
// @Failure 500 {object} GenericMessage
// @Failure 401
// @Router /heartrate/list [get]
func (h *HeartRateHandler) listHeartRate(w http.ResponseWriter, r *http.Request) {
	urlMatches := HeartRateListRgx.FindStringSubmatch(r.URL.String())

	if len(urlMatches) != 3 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, HeartRateListRgx.String())
		handleError(w, http.StatusInternalServerError, "Issue parsing URL")
		return
	}

	queryType := urlMatches[1]
	queryToken := urlMatches[2]

	InfoLog.Printf("URL directive match '%s'\n", queryType)
	InfoLog.Printf("URL token match '%s'\n", queryToken)

	params := austinapi_db.GetHeartRatesParams{
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

	results, err := ApiDatabase.GetHeartRates(DatabaseContext, params)
	if err != nil {
		ErrorLog.Printf("error getting list of heart rates: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(results) < 1 {
		ErrorLog.Printf("no heart rate results from database with '%s' token '%s'", queryType, queryToken)
		handleError(w, http.StatusNotFound, "no results found")
		return
	}

	heartrates := HeartRates{
		Data:      results,
		NextToken: params.RowLimit + params.RowOffset,
	}

	jsonBytes, err := json.Marshal(heartrates)
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
