package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func create(w http.ResponseWriter, r *http.Request) error {
	// TODO: Make the scraper_name optional
	// TODO: Do we make sure that there is only one scraper_name used?
	scraperName := r.URL.Query()["scraper_name"][0]

	createResult, err := commandCreate(scraperName)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createResult)
	return nil
}

func getApp(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return commandGetApp(runName, w)
}

func putApp(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return commandPutApp(r.Body, r.ContentLength, runName)
}

func getCache(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return commandGetCache(runName, w)
}

func putCache(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return commandPutCache(r.Body, r.ContentLength, runName)
}

func getOutput(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return commandGetOutput(runName, w)
}

func putOutput(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return commandPutOutput(r.Body, r.ContentLength, runName)
}

func getExitData(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return commandGetExitData(runName, w)
}

func putExitData(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return commandPutExitData(r.Body, r.ContentLength, runName)
}

func start(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]

	// TODO: If json is not of the right form return an error code that isn't 500
	decoder := json.NewDecoder(r.Body)
	var l startBody
	err := decoder.Decode(&l)
	if err != nil {
		return err
	}

	// TODO: If the scraper has already been started let the user know rather than 500'ing
	return commandStart(runName, l)
}

func getLogs(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]

	podLogs, err := commandGetLogs(runName)
	if err != nil {
		return err
	}
	defer podLogs.Close()

	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("Couldn't access the flusher")
	}

	scanner := bufio.NewScanner(podLogs)
	for scanner.Scan() {
		fmt.Fprintln(w, scanner.Text())
		flusher.Flush()
	}
	return scanner.Err()
}

func createLogs(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]

	decoder := json.NewDecoder(r.Body)
	var l logMessage
	err := decoder.Decode(&l)
	if err != nil {
		return err
	}
	return commandCreateLog(runName, l)
}

func delete(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]

	return commandDelete(runName)
}

func whoAmI(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintln(w, "Hello from Clay!")
	return nil
}

// Middleware that logs the request uri
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func extractBearerToken(header http.Header) (string, error) {
	const bearerPrefix = "Bearer "
	authHeader := header.Get("Authorization")

	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", errors.New("Expected Authorization header with bearer token")
	}
	return authHeader[len(bearerPrefix):], nil
}

// Middleware function, which will be called for each request
func authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runName := mux.Vars(r)["id"]
		runToken, err := extractBearerToken(r.Header)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		clientset, err := getClientSet()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		actualRunToken, err := actualRunToken(clientset, runName)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if runToken != actualRunToken {
			log.Println("Incorrect run token")
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type appHandler func(http.ResponseWriter, *http.Request) error

// Error handling
func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	log.Println("Clay is ready and waiting.")
	router := mux.NewRouter().StrictSlash(true)

	router.Handle("/", appHandler(whoAmI))
	router.Handle("/runs", appHandler(create)).Methods("POST")

	authenticatedRouter := router.PathPrefix("/runs/{id}").Subrouter()
	authenticatedRouter.Handle("/app", appHandler(getApp)).Methods("GET")
	authenticatedRouter.Handle("/app", appHandler(putApp)).Methods("PUT")
	authenticatedRouter.Handle("/cache", appHandler(getCache)).Methods("GET")
	authenticatedRouter.Handle("/cache", appHandler(putCache)).Methods("PUT")
	authenticatedRouter.Handle("/output", appHandler(getOutput)).Methods("GET")
	authenticatedRouter.Handle("/output", appHandler(putOutput)).Methods("PUT")
	authenticatedRouter.Handle("/exit-data", appHandler(getExitData)).Methods("GET")
	authenticatedRouter.Handle("/exit-data", appHandler(putExitData)).Methods("PUT")
	authenticatedRouter.Handle("/start", appHandler(start)).Methods("POST")
	authenticatedRouter.Handle("/logs", appHandler(getLogs)).Methods("GET")
	authenticatedRouter.Handle("/logs", appHandler(createLogs)).Methods("POST")
	authenticatedRouter.Handle("", appHandler(delete)).Methods("DELETE")
	authenticatedRouter.Use(authenticate)
	router.Use(logRequests)

	log.Fatal(http.ListenAndServe(":8080", router))
}
