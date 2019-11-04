package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"github.com/openaustralia/morph-ng/internal/commands"
)

func create(w http.ResponseWriter, r *http.Request) error {
	values := r.URL.Query()["name_prefix"]
	namePrefix := ""
	if len(values) > 0 {
		namePrefix = values[0]
	}

	createResult, err := app.CreateRun(namePrefix)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createResult)
	return nil
}

func getApp(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	w.Header().Set("Content-Type", "application/gzip")
	reader, err := app.GetApp(runName)
	if err != nil {
		// Returns 404 if there is no app
		if app.BlobStore.IsNotExist(err) {
			http.NotFound(w, r)
			return nil
		}
		return err
	}
	_, err = io.Copy(w, reader)
	return err
}

func putApp(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return app.PutApp(r.Body, r.ContentLength, runName)
}

func getCache(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	reader, err := app.GetCache(runName)
	if err != nil {
		// Returns 404 if there is no cache
		if app.BlobStore.IsNotExist(err) {
			http.NotFound(w, r)
			return nil
		}
		return err
	}
	w.Header().Set("Content-Type", "application/gzip")
	_, err = io.Copy(w, reader)
	return err
}

func putCache(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return app.PutCache(r.Body, r.ContentLength, runName)
}

func getOutput(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	reader, err := app.GetOutput(runName)
	if err != nil {
		// Returns 404 if there is no output
		if app.BlobStore.IsNotExist(err) {
			http.NotFound(w, r)
			return nil
		}
		return err
	}
	_, err = io.Copy(w, reader)
	return err
}

func putOutput(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return app.PutOutput(r.Body, r.ContentLength, runName)
}

func getExitData(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	reader, err := app.GetExitData(runName)
	if err != nil {
		// Returns 404 if there is no exit data
		if app.BlobStore.IsNotExist(err) {
			http.NotFound(w, r)
			return nil
		}
		return err
	}
	_, err = io.Copy(w, reader)
	return err
}

func putExitData(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return app.PutExitData(r.Body, r.ContentLength, runName)
}

type envVariable struct {
	Name  string
	Value string
}

type startBody struct {
	Output   string
	Env      []envVariable
	Callback callback
}

type callback struct {
	URL string
}

func start(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]

	decoder := json.NewDecoder(r.Body)
	var l startBody
	err := decoder.Decode(&l)
	if err != nil {
		return err
	}

	env := make(map[string]string)
	for _, keyvalue := range l.Env {
		env[keyvalue.Name] = keyvalue.Value
	}

	return app.StartRun(runName, l.Output, env, l.Callback.URL)
}

func getEvents(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	w.Header().Set("Content-Type", "application/ld+json")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("Couldn't access the flusher")
	}

	var id = "0"
	for {
		newID, jsonString, finished, err := app.GetEvent(runName, id)
		id = newID
		if err != nil {
			return err
		}
		if finished {
			break
		}
		fmt.Fprintln(w, jsonString)
		flusher.Flush()
	}
	return nil
}

func createEvents(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]

	// Read json message as is into a string
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return app.CreateEvent(runName, string(buf))
}

func delete(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]

	return app.DeleteRun(runName)
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

		actualRunToken, err := app.JobDispatcher.GetToken(runName)

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

// Our global state
var app *commands.App

func init() {
}

func main() {
	// Show the source of the error with the standard logger. Don't show date & time
	log.SetFlags(log.Lshortfile)

	var err error
	app, err = commands.New()
	if err != nil {
		log.Fatal(err)
	}

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
	authenticatedRouter.Handle("/events", appHandler(getEvents)).Methods("GET")
	authenticatedRouter.Handle("/events", appHandler(createEvents)).Methods("POST")
	authenticatedRouter.Handle("", appHandler(delete)).Methods("DELETE")
	authenticatedRouter.Use(authenticate)
	router.Use(logRequests)

	log.Fatal(http.ListenAndServe(":8080", router))
}
