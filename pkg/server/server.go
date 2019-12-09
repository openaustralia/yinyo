package server

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
	"github.com/openaustralia/yinyo/internal/commands"
	"github.com/openaustralia/yinyo/pkg/event"
)

func (server *Server) create(w http.ResponseWriter, r *http.Request) error {
	values := r.URL.Query()["name_prefix"]
	namePrefix := ""
	if len(values) > 0 {
		namePrefix = values[0]
	}

	createResult, err := server.app.CreateRun(namePrefix)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createResult)
	return nil
}

func (server *Server) getApp(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	w.Header().Set("Content-Type", "application/gzip")
	reader, err := server.app.GetApp(runName)
	if err != nil {
		// Returns 404 if there is no app
		if server.app.BlobStore.IsNotExist(err) {
			http.NotFound(w, r)
			return nil
		}
		return err
	}
	_, err = io.Copy(w, reader)
	return err
}

func (server *Server) putApp(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return server.app.PutApp(r.Body, r.ContentLength, runName)
}

func (server *Server) getCache(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	reader, err := server.app.GetCache(runName)
	if err != nil {
		// Returns 404 if there is no cache
		if server.app.BlobStore.IsNotExist(err) {
			http.NotFound(w, r)
			return nil
		}
		return err
	}
	w.Header().Set("Content-Type", "application/gzip")
	_, err = io.Copy(w, reader)
	return err
}

func (server *Server) putCache(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return server.app.PutCache(r.Body, r.ContentLength, runName)
}

func (server *Server) getOutput(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	reader, err := server.app.GetOutput(runName)
	if err != nil {
		// Returns 404 if there is no output
		if server.app.BlobStore.IsNotExist(err) {
			http.NotFound(w, r)
			return nil
		}
		return err
	}
	_, err = io.Copy(w, reader)
	return err
}

func (server *Server) putOutput(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return server.app.PutOutput(r.Body, r.ContentLength, runName)
}

func (server *Server) getExitData(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	reader, err := server.app.GetExitData(runName)
	if err != nil {
		// Returns 404 if there is no exit data
		if server.app.BlobStore.IsNotExist(err) {
			http.NotFound(w, r)
			return nil
		}
		return err
	}
	_, err = io.Copy(w, reader)
	return err
}

func (server *Server) putExitData(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	return server.app.PutExitData(r.Body, r.ContentLength, runName)
}

type envVariable struct {
	Name  string
	Value string
}

// TODO: Remove duplication with client code
type startBody struct {
	Output   string
	Env      []envVariable
	Callback callback
}

type callback struct {
	URL string
}

func (server *Server) start(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]

	decoder := json.NewDecoder(r.Body)
	var l startBody
	err := decoder.Decode(&l)
	if err != nil {
		return newHTTPError(err, http.StatusBadRequest, "JSON in body not correctly formatted")
	}

	env := make(map[string]string)
	for _, keyvalue := range l.Env {
		env[keyvalue.Name] = keyvalue.Value
	}

	return server.app.StartRun(runName, l.Output, env, l.Callback.URL)
}

func (server *Server) getEvents(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]
	lastID := mux.Vars(r)["last-id"]
	if lastID == "" {
		lastID = "0"
	}
	w.Header().Set("Content-Type", "application/ld+json")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return errors.New("Couldn't access the flusher")
	}

	events := server.app.GetEvents(runName, lastID)
	enc := json.NewEncoder(w)
	for events.More() {
		e, err := events.Next()
		if err != nil {
			return err
		}
		err = enc.Encode(e)
		if err != nil {
			return err
		}
		flusher.Flush()
	}
	return nil
}

func (server *Server) createEvent(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]

	// Read json message as is into a string
	// TODO: Switch over to json decoder
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// Check the form of the JSON by interpreting it
	var event event.Event
	err = json.Unmarshal(buf, &event)
	if err != nil {
		return newHTTPError(err, http.StatusBadRequest, "JSON in body not correctly formatted")
	}

	return server.app.CreateEvent(runName, event)
}

func (server *Server) delete(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]

	return server.app.DeleteRun(runName)
}

func (server *Server) whoAmI(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintln(w, "Hello from Yinyo!")
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
func (server *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runName := mux.Vars(r)["id"]
		runToken, err := extractBearerToken(r.Header)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		actualRunToken, err := server.app.JobDispatcher.GetToken(runName)

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
		err2, ok := err.(clientError)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body, err := err2.ResponseBody()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		status, headers := err2.ResponseHeaders()
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}

// Server holds the internal state for the server
type Server struct {
	app *commands.App
}

// Run runs the server. This function will block until the server quits
func Run() {
	var err error
	app, err := commands.New()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Yinyo is ready and waiting.")
	router := mux.NewRouter().StrictSlash(true)
	server := Server{app: app}

	router.Handle("/", appHandler(server.whoAmI))
	router.Handle("/runs", appHandler(server.create)).Methods("POST")

	authenticatedRouter := router.PathPrefix("/runs/{id}").Subrouter()
	authenticatedRouter.Handle("/app", appHandler(server.getApp)).Methods("GET")
	authenticatedRouter.Handle("/app", appHandler(server.putApp)).Methods("PUT")
	authenticatedRouter.Handle("/cache", appHandler(server.getCache)).Methods("GET")
	authenticatedRouter.Handle("/cache", appHandler(server.putCache)).Methods("PUT")
	authenticatedRouter.Handle("/output", appHandler(server.getOutput)).Methods("GET")
	authenticatedRouter.Handle("/output", appHandler(server.putOutput)).Methods("PUT")
	authenticatedRouter.Handle("/exit-data", appHandler(server.getExitData)).Methods("GET")
	authenticatedRouter.Handle("/exit-data", appHandler(server.putExitData)).Methods("PUT")
	authenticatedRouter.Handle("/start", appHandler(server.start)).Methods("POST")
	authenticatedRouter.Handle("/events", appHandler(server.getEvents)).Methods("GET")
	authenticatedRouter.Handle("/events", appHandler(server.createEvent)).Methods("POST")
	authenticatedRouter.Handle("", appHandler(server.delete)).Methods("DELETE")
	authenticatedRouter.Use(server.authenticate)
	router.Use(logRequests)

	log.Fatal(http.ListenAndServe(":8080", router))
}
