package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/austinmoody/austinapi_db/austinapi_db"
	"github.com/jackc/pgx/v5"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	ReadyScoreRgxId   *regexp.Regexp
	ReadyScoreListRgx *regexp.Regexp
	ReadyScoreRgxDate *regexp.Regexp
)

type ReadyScoreHandler struct{}

type ReadyScores struct {
	Data      []austinapi_db.Readyscore `json:"data"`
	NextToken int32                     `json:"next_token"`
}

func init() {
	ReadyScoreRgxId = regexp.MustCompile(`^/readyscore/id/([0-9]+)$`)
	ReadyScoreListRgx = regexp.MustCompile(`^/readyscore/list(?:\?(next_token)=([0-9]+))?$`)
	ReadyScoreRgxDate = regexp.MustCompile(`^/readyscore/date/([0-9]{4}-[0-9]{2}-[0-9]{2})$`)
}

func (h *ReadyScoreHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet && ReadyScoreListRgx.MatchString(r.URL.String()):
		h.ListReadyScore(w, r)
	case r.Method == http.MethodGet && ReadyScoreRgxId.MatchString(r.URL.String()):
		h.GetReadyScore(w, r)
	case r.Method == http.MethodGet && ReadyScoreRgxDate.MatchString(r.URL.String()):
		h.GetReadyScoreByDate(w, r)
	default:
		handleError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// @Summary Get ready score information by ID
// @Security ApiKeyAuth
// @Description Retrieves ready score information with specified ID
// @Tags readyscore
// @Accept json
// @Produce json
// @Param id path string true "Ready Score ID"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} austinapi_db.Readyscore
// @Failure 500 {object} GenericMessage
// @Failure 404 {object} GenericMessage
// @Failure 401
// @Router /readyscore/id/{id} [get]
func (h *ReadyScoreHandler) GetReadyScore(w http.ResponseWriter, r *http.Request) {
	idMatches := ReadyScoreRgxId.FindStringSubmatch(r.URL.String())

	if len(idMatches) < 2 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, ReadyScoreRgxId.String())
		handleError(w, http.StatusInternalServerError, "Issue parsing specified id")
		return
	}

	InfoLog.Printf("URL token match '%s'\n", idMatches[1])

	id, err := strconv.ParseInt(idMatches[1], 10, 64)
	if err != nil {
		ErrorLog.Printf("issue converting id to int64: %v", err)
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

	result, err := apiDb.GetReadyScore(ctx, id)

	if err != nil {
		ErrorLog.Printf("error retrieving ready score with id '%d': %v", id, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(result) != 1 {
		InfoLog.Printf("ready score with id '%d' was not found in database", id)
		handleError(w, http.StatusNotFound, fmt.Sprintf("Ready Score not found with id %d", id))
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

// @Summary Get ready score information by date
// @Security ApiKeyAuth
// @Description Retrieves ready score information with specified date
// @Tags readyscore
// @Accept json
// @Produce json
// @Param date path string true "Date"
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} austinapi_db.Readyscore
// @Failure 500 {object} GenericMessage
// @Failure 404 {object} GenericMessage
// @Failure 401
// @Router /readyscore/date/{date} [get]
func (h *ReadyScoreHandler) GetReadyScoreByDate(w http.ResponseWriter, r *http.Request) {

	dateMatches := ReadyScoreRgxDate.FindStringSubmatch(r.URL.String())

	if len(dateMatches) < 2 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, ReadyScoreRgxDate.String())
		handleError(w, http.StatusInternalServerError, "Issue parsing specified date")
		return
	}

	InfoLog.Printf("URL token match '%s'\n", dateMatches[1])

	dateString := dateMatches[1]
	searchDate, err := time.Parse("2006-01-02", dateString)
	if err != nil {
		ErrorLog.Printf("Unable to parse '%s' to time.Time object: %v", dateString, err)
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

	result, err := apiDb.GetReadyScoreByDate(ctx, searchDate)

	if err != nil {
		ErrorLog.Printf("error retrieving ready score with date '%s': %v", dateString, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(result) != 1 {
		InfoLog.Printf("ready score with date '%s' was not found in database", dateString)
		handleError(w, http.StatusNotFound, fmt.Sprintf("Ready Score not found with date %s", dateString))
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

// @Summary Get list of ready score information
// @Security ApiKeyAuth
// @Description Retrieves list of ready score information in descending order by date
// @Description Specifying no query parameters pulls list starting with latest
// @Description Caller can then specify a next_token from previous calls to go
// @Description forward in the list of items.
// @Tags readyscore
// @Produce json
// @Param next_token query string false "next list search by next_token" Format(string)
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} ReadyScores
// @Failure 500 {object} GenericMessage
// @Failure 401
// @Router /readyscore/list [get]
func (h *ReadyScoreHandler) ListReadyScore(w http.ResponseWriter, r *http.Request) {
	urlMatches := ReadyScoreListRgx.FindStringSubmatch(r.URL.String())

	if len(urlMatches) != 3 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, ReadyScoreListRgx.String())
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

	params := austinapi_db.GetReadyScoresParams{
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

	results, err := apiDb.GetReadyScores(ctx, params)
	if err != nil {
		ErrorLog.Printf("error getting list of ready scores: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(results) < 1 {
		ErrorLog.Printf("no ready score results from database with '%s' token '%s'", queryType, queryToken)
		handleError(w, http.StatusNotFound, "no results found")
		return
	}

	readyScores := ReadyScores{
		Data:      results,
		NextToken: params.RowLimit + params.RowOffset,
	}

	jsonBytes, err := json.Marshal(readyScores)
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
