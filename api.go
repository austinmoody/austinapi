package main

import (
	"encoding/json"
	"github.com/austinmoody/austinapi_db/austinapi_db"
	"log"
	"net/http"
	"regexp"
	"time"
)

var (
	SleepRegex       = regexp.MustCompile(`^/sleep/*$`)
	SleepRegexWithID = regexp.MustCompile(`^/sleep/([0-9]+)$`)
)

func main() {

	mux := http.NewServeMux()

	mux.Handle("/sleep", &SleepHandler{})
	mux.Handle("/sleep/", &SleepHandler{})

	http.ListenAndServe(":8080", mux)

}

type SleepHandler struct{}

func (h *SleepHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && SleepRegex.MatchString(r.URL.Path):
		h.ListSleep(w, r)
	case r.Method == http.MethodGet && SleepRegexWithID.MatchString(r.URL.Path):
		h.GetSleep(w, r)
	default:
		return
	}
}

func (sh *SleepHandler) GetSleep(w http.ResponseWriter, r *http.Request) {
	sleep := austinapi_db.Sleep{
		ID:               123,
		Date:             time.Now(), // TODO need to have "Day" object in db, aggregator, here, etc...
		Rating:           65,
		TotalSleep:       34345345,
		DeepSleep:        23423,
		LightSleep:       22343,
		RemSleep:         3535,
		CreatedTimestamp: time.Now(),
		UpdatedTimestamp: time.Now(),
	}

	jsonBytes, err := json.Marshal(sleep)
	if err != nil {
		log.Fatalf("%v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (sh *SleepHandler) ListSleep(w http.ResponseWriter, r *http.Request) {

}

//package main
//
//import (
//	"encoding/json"
//	"fmt"
//	"log"
//	"net/http"
//)
//
//// Item represents a generic item in the API
//type Item struct {
//	ID   string `json:"id"`
//	Name string `json:"name"`
//}
//
//var items = map[string]Item{}
//
//func main() {
//	http.HandleFunc("/items", handleItems)
//	http.HandleFunc("/items/", handleItem)
//
//	fmt.Println("Server listening on port 8080...")
//	log.Fatal(http.ListenAndServe(":8080", nil))
//}
//
//func handleItems(w http.ResponseWriter, r *http.Request) {
//	switch r.Method {
//	case http.MethodGet:
//		getItems(w, r)
//	case http.MethodPost:
//		createItem(w, r)
//	default:
//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//	}
//}
//
//func getItems(w http.ResponseWriter, r *http.Request) {
//	itemsList := make([]Item, 0, len(items))
//	for _, v := range items {
//		itemsList = append(itemsList, v)
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	json.NewEncoder(w).Encode(itemsList)
//}
//
//func createItem(w http.ResponseWriter, r *http.Request) {
//	var newItem Item
//	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
//		http.Error(w, "Invalid request body", http.StatusBadRequest)
//		return
//	}
//
//	items[newItem.ID] = newItem
//	w.WriteHeader(http.StatusCreated)
//}
//
//func handleItem(w http.ResponseWriter, r *http.Request) {
//	id := r.URL.Path[len("/items/"):]
//	switch r.Method {
//	case http.MethodGet:
//		getItem(w, r, id)
//	default:
//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//	}
//}
//
//func getItem(w http.ResponseWriter, r *http.Request, id string) {
//	item, found := items[id]
//	if !found {
//		http.Error(w, "Item not found", http.StatusNotFound)
//		return
//	}
//
//	w.Header().Set("Content-Type", "application/json")
//	json.NewEncoder(w).Encode(item)
//}
