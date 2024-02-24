package main

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"github.com/austinmoody/austinapi_db/austinapi_db"
//	"github.com/jackc/pgx/v5"
//	"net/http"
//	"regexp"
//	"time"
//)
//
//var (
//	ReadyScoreRgxId   *regexp.Regexp
//	ReadyScoreListRgx *regexp.Regexp
//	ReadyScoreRgxDate *regexp.Regexp
//)
//
//type ReadyScoreHandler struct{}
//
//type ReadyScore struct {
//	ID               string
//	Date             time.Time
//	Score            int
//	CreatedTimestamp time.Time
//	UpdatedTimestamp time.Time
//}
//
//type ReadyScores struct {
//	Data          []ReadyScore `json:"data"`
//	NextToken     string       `json:"next_token"`
//	PreviousToken string       `json:"previous_token"`
//}
//
//type ReadyScoreResult austinapi_db.Readyscore
//type ReadyScoresResult []austinapi_db.GetReadyScoresRow
//type ReadyScoresResultItem austinapi_db.GetReadyScoresRow
//
//func (rsi ReadyScoreResult) ToReadyScore() ReadyScore {
//
//	var rs ReadyScore
//
//	rs.ID, _ = IdHasher.Encode([]uint64{uint64(rsi.ID)})
//	rs.Date = rsi.Date
//	rs.Score = rsi.Score
//	rs.CreatedTimestamp = rsi.CreatedTimestamp
//	rs.UpdatedTimestamp = rsi.UpdatedTimestamp
//
//	return rs
//}
//
//func (rsri ReadyScoresResultItem) ToReadyScore() ReadyScore {
//	var rs ReadyScore
//
//	rs.ID, _ = IdHasher.Encode([]uint64{uint64(rsri.ID)})
//	rs.Date = rsri.Date
//	rs.Score = rsri.Score
//	rs.CreatedTimestamp = rsri.CreatedTimestamp
//	rs.UpdatedTimestamp = rsri.UpdatedTimestamp
//
//	return rs
//}
//
//func (rsr ReadyScoresResult) ToReadyScores() ReadyScores {
//	var scores []ReadyScore
//	for _, result := range rsr {
//		scores = append(scores, ReadyScoresResultItem(result).ToReadyScore())
//	}
//
//	return ReadyScores{
//		Data:          scores,
//		NextToken:     rsr.GetNextToken(),
//		PreviousToken: rsr.GetPreviousToken(),
//	}
//}
//
//func (rsr ReadyScoresResult) GetNextToken() string {
//	var nextToken string
//
//	nextTokenInt := rsr[len(rsr)-1].NextID
//	if nextTokenInt < 1 {
//		nextToken = ""
//	} else {
//		id, _ := IdHasher.Encode([]uint64{uint64(nextTokenInt)})
//		nextToken = id
//	}
//
//	return nextToken
//}
//
//func (rsr ReadyScoresResult) GetPreviousToken() string {
//	var previousToken string
//	previousTokenInt := rsr[0].PreviousID
//	if previousTokenInt < 1 {
//		previousToken = ""
//	} else {
//		id, _ := IdHasher.Encode([]uint64{uint64(previousTokenInt)})
//		previousToken = id
//	}
//
//	return previousToken
//}
//
//func init() {
//	ReadyScoreRgxId = regexp.MustCompile(fmt.Sprintf(`^/readyscore/id/([0-9a-zA-Z]{%s})$`, SqidLength))
//	ReadyScoreListRgx = regexp.MustCompile(fmt.Sprintf(`^/readyscore/list(?:\?(next_token|previous_token)=([0-9a-zA-Z]{%s}))?$`, SqidLength))
//	ReadyScoreRgxDate = regexp.MustCompile(`^/readyscore/date/([0-9]{4}-[0-9]{2}-[0-9]{2})$`)
//}
//
//func (h *ReadyScoreHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	w.Header().Set("Content-Type", "application/json")
//
//	switch {
//	case r.Method == http.MethodGet && ReadyScoreListRgx.MatchString(r.URL.String()):
//		h.ListReadyScore(w, r)
//	case r.Method == http.MethodGet && ReadyScoreRgxId.MatchString(r.URL.String()):
//		h.GetReadyScore(w, r)
//	case r.Method == http.MethodGet && ReadyScoreRgxDate.MatchString(r.URL.String()):
//		h.GetReadyScoreByDate(w, r)
//	default:
//		handleError(w, http.StatusMethodNotAllowed, "Method not allowed")
//	}
//}
//
//func (h *ReadyScoreHandler) GetReadyScore(w http.ResponseWriter, r *http.Request) {
//	idMatches := ReadyScoreRgxId.FindStringSubmatch(r.URL.String())
//
//	if len(idMatches) < 2 {
//		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, ReadyScoreRgxId.String())
//		handleError(w, http.StatusInternalServerError, "Issue parsing specified id")
//		return
//	}
//
//	InfoLog.Printf("URL token match '%s'\n", idMatches[1])
//
//	sqid := idMatches[1]
//
//	connStr := GetDatabaseConnectionString()
//	ctx := context.Background()
//
//	conn, err := pgx.Connect(ctx, connStr)
//	if err != nil {
//		ErrorLog.Printf("DB Connection error: %v", err)
//		handleError(w, http.StatusInternalServerError, "Internal Error")
//		return
//	}
//	defer conn.Close(ctx)
//
//	apiDb := austinapi_db.New(conn)
//
//	idFromSqid := GetIdFromToken(sqid)
//
//	result, err := apiDb.GetReadyScore(ctx, idFromSqid)
//
//	if err != nil {
//		ErrorLog.Printf("error retrieving ready score with id '%v': %v", idFromSqid, err)
//		handleError(w, http.StatusInternalServerError, "Internal Error")
//		return
//	}
//
//	if len(result) != 1 {
//		InfoLog.Printf("ready score with id '%v was not found in database", idFromSqid)
//		handleError(w, http.StatusNotFound, fmt.Sprintf("Ready Score not found with id %s", sqid))
//		return
//	}
//
//	jsonBytes, err := json.Marshal(ReadyScoreResult(result[0]).ToReadyScore())
//	if err != nil {
//		ErrorLog.Printf("error marshaling JSON response: %v", err)
//		handleError(w, http.StatusInternalServerError, "Internal Error")
//		return
//	}
//
//	w.WriteHeader(http.StatusOK)
//	_, err = w.Write(jsonBytes)
//	if err != nil {
//		ErrorLog.Printf("error writing http response: %v\n", err)
//	}
//}
//
//func (h *ReadyScoreHandler) GetReadyScoreByDate(w http.ResponseWriter, r *http.Request) {}
//
//func (h *ReadyScoreHandler) ListReadyScore(w http.ResponseWriter, r *http.Request) {
//	urlMatches := ReadyScoreListRgx.FindStringSubmatch(r.URL.String())
//
//	if len(urlMatches) != 3 {
//		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, ReadyScoreListRgx.String())
//		handleError(w, http.StatusInternalServerError, "Issue parsing URL")
//		return
//	}
//
//	queryType := urlMatches[1]
//	queryToken := urlMatches[2]
//
//	InfoLog.Printf("URL directive match '%s'\n", queryType)
//	InfoLog.Printf("URL token match '%s'\n", queryToken)
//
//	connStr := GetDatabaseConnectionString()
//	ctx := context.Background()
//
//	conn, err := pgx.Connect(ctx, connStr)
//	if err != nil {
//		ErrorLog.Printf("database connection error: %v", err)
//		handleError(w, http.StatusInternalServerError, "Internal Error")
//		return
//	}
//	defer conn.Close(ctx)
//
//	apiDb := austinapi_db.New(conn)
//
//	params := austinapi_db.GetReadyScoresParams{
//		QueryType: "",
//		InputID:   0,
//		RowLimit:  ListRowLimit,
//	}
//
//	switch queryType {
//	case "next_token":
//		params.QueryType = "NEXT"
//		params.InputID = GetIdFromToken(queryToken)
//	case "previous_token":
//		params.QueryType = "PREVIOUS"
//		params.InputID = GetIdFromToken(queryToken)
//	}
//
//	results, err := apiDb.GetReadyScores(ctx, params)
//	if err != nil {
//		ErrorLog.Printf("error getting list of ready scores: %v", err)
//		handleError(w, http.StatusInternalServerError, "Internal Error")
//		return
//	}
//
//	if len(results) < 1 {
//		ErrorLog.Printf("no ready score results from database with '%' token '%s'", queryType, queryToken)
//		handleError(w, http.StatusNotFound, "no results found")
//		return
//	}
//
//	jsonBytes, err := json.Marshal(ReadyScoresResult(results).ToReadyScores())
//	if err != nil {
//		ErrorLog.Printf("error marshaling JSON response: %v", err)
//		handleError(w, http.StatusInternalServerError, "Internal Error")
//		return
//	}
//
//	w.WriteHeader(http.StatusOK)
//	_, err = w.Write(jsonBytes)
//	if err != nil {
//		ErrorLog.Printf("error writing http response: %v", err)
//	}
//}
