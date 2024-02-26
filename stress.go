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
	StressRgxId   *regexp.Regexp
	StressListRgx *regexp.Regexp
	StressRgxDate *regexp.Regexp
)

type StressHandler struct{}

type Stresses struct {
	Data      []austinapi_db.Stress `json:"data"`
	NextToken int32                 `json:"next_token"`
}

func init() {
	StressRgxId = regexp.MustCompile(`^/stress/id/([0-9]+)$`)
	StressListRgx = regexp.MustCompile(`^/stress/list(?:\?(next_token)=([0-9]+))?$`)
	StressRgxDate = regexp.MustCompile(`^/stress/date/([0-9]{4}-[0-9]{2}-[0-9]{2})$`)
}

func (h *StressHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet && StressListRgx.MatchString(r.URL.String()):
		h.listStress(w, r)
	case r.Method == http.MethodGet && StressRgxId.MatchString(r.URL.String()):
		h.getStress(w, r)
	case r.Method == http.MethodGet && StressRgxDate.MatchString(r.URL.String()):
		h.getStressByDate(w, r)
	default:
		handleError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// @Summary Get stress information by ID
// @Security ApiKeyAuth
// @Description Retrieves stress information with specified ID
// @Tags stress
// @Accept json
// @Produce json
// @Param id path string true "Stress ID"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} austinapi_db.Stress
// @Failure 500 {object} GenericMessage
// @Failure 404 {object} GenericMessage
// @Failure 401
// @Router /stress/id/{id} [get]
func (h *StressHandler) getStress(w http.ResponseWriter, r *http.Request) {

	id, err := getIdFromUrl(StressRgxId, r.URL)

	if err != nil {
		ErrorLog.Println(err)
		handleError(w, http.StatusInternalServerError, "Issue parsing id from URL")
		return
	}

	InfoLog.Printf("URL id match '%d'\n", id)

	result, err := ApiDatabase.GetStress(DatabaseContext, id)

	if err != nil {
		ErrorLog.Printf("error retrieving stress with id '%d': %v", id, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(result) != 1 {
		InfoLog.Printf("stress with id '%d' was not found in database", id)
		handleError(w, http.StatusNotFound, fmt.Sprintf("Stress not found with id %d", id))
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

// @Summary Get stress information by date
// @Security ApiKeyAuth
// @Description Retrieves stress information with specified date
// @Tags stress
// @Accept json
// @Produce json
// @Param date path string true "Date"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} austinapi_db.Stress
// @Failure 500 {object} GenericMessage
// @Failure 404 {object} GenericMessage
// @Failure 401
// @Router /stress/date/{date} [get]
func (h *StressHandler) getStressByDate(w http.ResponseWriter, r *http.Request) {

	dateMatches := StressRgxDate.FindStringSubmatch(r.URL.String())

	if len(dateMatches) < 2 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, StressRgxDate.String())
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

	result, err := ApiDatabase.GetStressByDate(DatabaseContext, date)

	if err != nil {
		ErrorLog.Printf("error retrieving stress with date '%v': %v", dateString, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(result) != 1 {
		InfoLog.Printf("stress with date '%s' was not found in database", dateString)
		handleError(w, http.StatusNotFound, fmt.Sprintf("Stress not found with date %s", dateString))
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

// @Summary Get list of stress information
// @Security ApiKeyAuth
// @Description Retrieves list of stress information in descending order by date
// @Description Specifying no query parameters pulls list starting with latest
// @Description Caller can then specify a next_token from previous calls to go
// @Description forward in the list of items.
// @Tags stress
// @Produce json
// @Param next_token query string false "next list search by next_token" Format(string)
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} Stresses
// @Failure 500 {object} GenericMessage
// @Failure 401
// @Router /stress/list [get]
func (h *StressHandler) listStress(w http.ResponseWriter, r *http.Request) {
	urlMatches := StressListRgx.FindStringSubmatch(r.URL.String())

	if len(urlMatches) != 3 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, StressListRgx.String())
		handleError(w, http.StatusInternalServerError, "Issue parsing URL")
		return
	}

	queryType := urlMatches[1]
	queryToken := urlMatches[2]

	InfoLog.Printf("URL directive match '%s'\n", queryType)
	InfoLog.Printf("URL token match '%s'\n", queryToken)

	params := austinapi_db.GetStressesParams{
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

	results, err := ApiDatabase.GetStresses(DatabaseContext, params)
	if err != nil {
		ErrorLog.Printf("error getting list of stress: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(results) < 1 {
		ErrorLog.Printf("no stress results from database with '%s' token '%s'", queryType, queryToken)
		handleError(w, http.StatusNotFound, "no results found")
		return
	}

	stresses := Stresses{
		Data:      results,
		NextToken: params.RowLimit + params.RowOffset,
	}

	jsonBytes, err := json.Marshal(stresses)
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
