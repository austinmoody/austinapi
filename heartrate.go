package main

import (
	"github.com/austinmoody/austinapi_db/austinapi_db"
	"net/http"
	"regexp"
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

func (h *HeartRateHandler) getHeartRate(w http.ResponseWriter, r *http.Request) {}

func (h *HeartRateHandler) getHeartRateByDate(w http.ResponseWriter, r *http.Request) {}

func (h *HeartRateHandler) listHeartRate(w http.ResponseWriter, r *http.Request) {}
