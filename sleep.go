package main

import (
	"context"
	"encoding/json"
	"github.com/austinmoody/austinapi_db/austinapi_db"
	"github.com/jackc/pgx/v5"
	"github.com/sqids/sqids-go"
	"log"
	"net/http"
	"os"
	"regexp"
)

var (
	SleepRgx   = regexp.MustCompile(`^/sleep/*$`)
	SleepRgxId = regexp.MustCompile(`^/sleep/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$`)
	//SleepRgxNextToken = regexp.MustCompile(`^/sleep\?(next\_token)\=([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$`)
	SleepRgxNextToken = regexp.MustCompile(`^/sleep\?(next\_token)\=([0-9a-zA-Z]{10})$`)

	InfoLog  *log.Logger
	ErrorLog *log.Logger
)

type SleepHandler struct{}

// TODO should ListSleepRowToSleep type conversions live in DB or here?
type Sleeps struct {
	Data          []austinapi_db.Sleep `json:"data"`
	NextToken     string               `json:"next_token"`
	PreviousToken string               `json:"previous_token"`
}

func (s *Sleeps) CreateFromDbListSleepRow(rows []austinapi_db.ListSleepRow) {
	s.Data = austinapi_db.ListSleepRowToSleep(rows)

	var previousId string
	if rows[0].PreviousID < 1 {
		previousId = ""
	} else {
		s, _ := sqids.New(sqids.Options{
			MinLength: 10,
		})
		id, _ := s.Encode([]uint64{uint64(rows[0].PreviousID)})
		previousId = id
	}

	var nextId string
	if rows[len(rows)-1].NextID < 1 {
		nextId = ""
	} else {
		s, _ := sqids.New(sqids.Options{
			MinLength: 10,
		})
		id, _ := s.Encode([]uint64{uint64(rows[len(rows)-1].NextID)})
		nextId = id
	}

	s.NextToken = nextId
	s.PreviousToken = previousId

}

func init() {
	InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func (h *SleepHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet && SleepRgx.MatchString(r.URL.String()):
		h.ListSleep(w, r)
	case r.Method == http.MethodGet && SleepRgxId.MatchString(r.URL.String()):
		h.GetSleep(w, r)
	case r.Method == http.MethodGet && SleepRgxNextToken.MatchString(r.URL.String()):
		h.GetSleepNextToken(w, r)
	default:
		handleError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}

}

func handleError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(GenericMessage{Message: message})
}

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

func (h *SleepHandler) GetSleepNextToken(w http.ResponseWriter, r *http.Request) {
	urlMatches := SleepRgxNextToken.FindStringSubmatch(r.URL.String())

	if len(urlMatches) != 3 {
		ErrorLog.Printf("error regex parsing url '%s' with regex '%s'", r.URL.Path, SleepRgxNextToken.String())
		handleError(w, http.StatusInternalServerError, "Issue parsing next_token and id")
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

	nextTokenSqid := urlMatches[2]
	nextTokenSlice := IdHasher.Decode(nextTokenSqid)
	if err != nil {
		ErrorLog.Printf("error decoding next_token: %v", err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}
	nextToken := nextTokenSlice[0]

	params := austinapi_db.ListSleepNextParams{
		ID:    int64(nextToken),
		Limit: 10,
	}
	sleeps, err := apiDb.ListSleepNext(ctx, params)
	if err != nil {
		ErrorLog.Printf("error retrieving next items with id '%d': %v", nextToken, err)
		handleError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	nextSleeps := austinapi_db.ListSleepNextRowToSleep(sleeps)

	var previousId string
	if sleeps[0].PreviousID < 1 {
		previousId = ""
	} else {
		id, _ := IdHasher.Encode([]uint64{uint64(sleeps[0].PreviousID)})
		previousId = id
	}

	var nextId string
	if sleeps[len(sleeps)-1].NextID < 1 {
		nextId = ""
	} else {
		id, _ := IdHasher.Encode([]uint64{uint64(sleeps[len(sleeps)-1].NextID)})
		nextId = id
	}

	sleepList := Sleeps{
		Data:          nextSleeps,
		NextToken:     nextId,
		PreviousToken: previousId,
	}

	// TODO - Return ID as sqid?
	jsonBytes, err := json.Marshal(sleepList)
	if err != nil {
		log.Fatalf("%v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *SleepHandler) ListSleep(w http.ResponseWriter, r *http.Request) {

	connStr := GetDatabaseConnectionString()
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("DB Connection error: %v", err)
	}
	defer conn.Close(ctx)

	apiDb := austinapi_db.New(conn)

	sleepsFromDb, err := apiDb.ListSleep(ctx, 10)
	if err != nil {
		log.Fatalf("Error getting list of sleep %v\n", err)
	}

	sleeps := Sleeps{}
	sleeps.CreateFromDbListSleepRow(sleepsFromDb)

	jsonBytes, err := json.Marshal(sleeps)
	if err != nil {
		log.Fatalf("error marshaling JSON response: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)

}
