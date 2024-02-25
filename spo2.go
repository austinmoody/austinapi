package main

import (
	"github.com/austinmoody/austinapi_db/austinapi_db"
	"net/http"
	"regexp"
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

func (h *Spo2Handler) getSpo2(w http.ResponseWriter, r *http.Request) {}

func (h *Spo2Handler) getSpo2ByDate(w http.ResponseWriter, r *http.Request) {}

func (h *Spo2Handler) listSpo2(w http.ResponseWriter, r *http.Request) {}
