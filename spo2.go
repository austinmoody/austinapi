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
	Spo2RgxId   *regexp.Regexp
	Spo2ListRgx *regexp.Regexp
	Spo2RgxDate *regexp.Regexp
)

type Spo2Handler struct{}

type Spo2s struct {
	Data      []austinapi_db.Spo2 `json:"data"`
	NextToken int32               `json:"next_token"`
}

func init() {
	Spo2RgxId = regexp.MustCompile(`^/spo2/id/([0-9]+)$`)
	Spo2ListRgx = regexp.MustCompile(`^/spo2/list(?:\?(next_token)=([0-9]+))?$`)
	Spo2RgxDate = regexp.MustCompile(`^/spo2/date/([0-9]{4}-[0-9]{2}-[0-9]{2})$`)
}

func (h *Spo2Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet && Spo2ListRgx.MatchString(r.URL.String()):
		h.listSpo2(w, r)
	case r.Method == http.MethodGet && Spo2RgxId.MatchString(r.URL.String()):
		h.getSpo2(w, r)
	case r.Method == http.MethodGet && Spo2RgxDate.MatchString(r.URL.String()):
		h.getSpo2ByDate(w, r)
	default:
		handleError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// @Summary Get Spo2 information by ID
// @Security ApiKeyAuth
// @Description Retrieves Spo2 information with specified ID
// @Tags spo2
// @Accept json
// @Produce json
// @Param id path string true "Spo2 ID"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} austinapi_db.Spo2
// @Failure 500 {object} GenericMessage
// @Failure 404 {object} GenericMessage
// @Failure 401
// @Router /spo2/id/{id} [get]
func (h *Spo2Handler) getSpo2(w http.ResponseWriter, r *http.Request) {
	id, err := getIdFromUrl(Spo2RgxId, r.URL)

	if err != nil {
		ErrorLog.Println(err)
		handleError(w, http.StatusInternalServerError, "Issue parsing id from URL")
		return
	}

	InfoLog.Printf("URL id match '%d'\n", id)

	result, err := ApiDatabase.GetSpo2(DatabaseContext, id)

	if err != nil {
		ErrorLog.Printf("error retrieving spo2 with id '%d': %v", id, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(result) != 1 {
		InfoLog.Printf("spo2 with id '%d' was not found in database", id)
		handleError(w, http.StatusNotFound, fmt.Sprintf("Spo2 not found with id %d", id))
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

// @Summary Get spo2 information by date
// @Security ApiKeyAuth
// @Description Retrieves spo2 information with specified date
// @Tags spo2
// @Accept json
// @Produce json
// @Param date path string true "Date"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} austinapi_db.Spo2
// @Failure 500 {object} GenericMessage
// @Failure 404 {object} GenericMessage
// @Failure 401
// @Router /spo2/date/{date} [get]
func (h *Spo2Handler) getSpo2ByDate(w http.ResponseWriter, r *http.Request) {
	dateMatches := Spo2RgxDate.FindStringSubmatch(r.URL.String())

	if len(dateMatches) < 2 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, Spo2RgxDate.String())
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

	result, err := ApiDatabase.GetSpo2ByDate(DatabaseContext, date)

	if err != nil {
		ErrorLog.Printf("error retrieving spo2 with date '%v': %v", dateString, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(result) != 1 {
		InfoLog.Printf("spo2 with date '%s' was not found in database", dateString)
		handleError(w, http.StatusNotFound, fmt.Sprintf("Spo2 not found with date %s", dateString))
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

// @Summary Get list of spo2 information
// @Security ApiKeyAuth
// @Description Retrieves list of spo2 information in descending order by date
// @Description Specifying no query parameters pulls list starting with latest
// @Description Caller can then specify a next_token from previous calls to go
// @Description forward in the list of items.
// @Tags spo2
// @Produce json
// @Param next_token query string false "next list search by next_token" Format(string)
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} Spo2s
// @Failure 500 {object} GenericMessage
// @Failure 401
// @Router /spo2/list [get]
func (h *Spo2Handler) listSpo2(w http.ResponseWriter, r *http.Request) {
	urlMatches := Spo2ListRgx.FindStringSubmatch(r.URL.String())

	if len(urlMatches) != 3 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, Spo2ListRgx.String())
		handleError(w, http.StatusInternalServerError, "Issue parsing URL")
		return
	}

	queryType := urlMatches[1]
	queryToken := urlMatches[2]

	InfoLog.Printf("URL directive match '%s'\n", queryType)
	InfoLog.Printf("URL token match '%s'\n", queryToken)

	params := austinapi_db.GetSpo2sParams{
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

	results, err := ApiDatabase.GetSpo2s(DatabaseContext, params)
	if err != nil {
		ErrorLog.Printf("error getting list of spo2: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(results) < 1 {
		ErrorLog.Printf("no spo2 results from database with '%s' token '%s'", queryType, queryToken)
		handleError(w, http.StatusNotFound, "no results found")
		return
	}

	spo2s := Spo2s{
		Data:      results,
		NextToken: params.RowLimit + params.RowOffset,
	}

	jsonBytes, err := json.Marshal(spo2s)
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
