package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// The body of the request should contain the tarred & gzipped code
// to be run
func run(w http.ResponseWriter, r *http.Request) {
	scraperID := mux.Vars(r)["id"]
	outputFilename := r.Header.Get("Clay-Scraper-Output")

	fmt.Fprintf(w, "scraperID: %s, outputFilename: %s", scraperID, outputFilename)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/scrapers/{id}/run", run).Methods("PUT")
	log.Fatal(http.ListenAndServe(":8080", router))
}
