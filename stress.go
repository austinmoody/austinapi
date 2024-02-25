package main

import (
	"github.com/austinmoody/austinapi_db/austinapi_db"
	"net/http"
	"regexp"
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

func (h *StressHandler) getStress(w http.ResponseWriter, r *http.Request) {}

func (h *StressHandler) getStressByDate(w http.ResponseWriter, r *http.Request) {}

func (h *StressHandler) listStress(w http.ResponseWriter, r *http.Request) {}
