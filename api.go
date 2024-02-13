package main

import (
	"github.com/sqids/sqids-go"
	"net/http"
)

var (
	IdHasher sqids.Sqids
)

func init() {
	s, _ := sqids.New(sqids.Options{
		MinLength: 10,
	})

	IdHasher = *s
}

func main() {

	mux := http.NewServeMux()

	mux.Handle("/sleep", &SleepHandler{})
	mux.Handle("/sleep/", &SleepHandler{})

	http.ListenAndServe(":8080", mux)

}
