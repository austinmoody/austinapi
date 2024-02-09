package main

import (
	"net/http"
)

func main() {

	mux := http.NewServeMux()

	mux.Handle("/sleep", &SleepHandler{})
	mux.Handle("/sleep/", &SleepHandler{})

	http.ListenAndServe(":8080", mux)

}
