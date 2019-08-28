package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

type createResult struct {
	RunName  string `json:"run_name"`
	RunToken string `json:"run_token"`
}

func create(w http.ResponseWriter, r *http.Request) error {
	// TODO: Make the scraper_name optional
	// TODO: Do we make sure that there is only one scraper_name used?
	scraperName := r.URL.Query()["scraper_name"][0]

	clientset, err := getClientSet()
	if err != nil {
		return err
	}

	runName, runToken, err := createSecret(clientset, scraperName)
	if err != nil {
		return err
	}

	createResult := createResult{
		RunName:  runName,
		RunToken: runToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createResult)
	return nil
}

func store(w http.ResponseWriter, r *http.Request, fileName string, fileExtension string) error {
	runName := mux.Vars(r)["id"]

	if r.Method == "GET" {
		return retrieveFromStore(runName, fileName, fileExtension, w)
	}
	return saveToStore(r.Body, r.ContentLength, runName, fileName, fileExtension)
}

// The body of the request should contain the tarred & gzipped code
func app(w http.ResponseWriter, r *http.Request) error {
	return store(w, r, "app", "tgz")
}

// The body of the request should contain the tarred & gzipped cache
func cache(w http.ResponseWriter, r *http.Request) error {
	return store(w, r, "cache", "tgz")
}

func output(w http.ResponseWriter, r *http.Request) error {
	return store(w, r, "output", "")
}

func exitData(w http.ResponseWriter, r *http.Request) error {
	return store(w, r, "exit-data", "json")
}

type startBody struct {
	Output string
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

	runOutput := l.Output

	clientset, err := getClientSet()
	if err != nil {
		return err
	}

	return createJob(clientset, runName, runOutput)
}

type logMessage struct {
	// TODO: Make the stream an enum
	Log, Stream string
}

func logs(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]

	clientset, err := getClientSet()
	if err != nil {
		return err
	}

	if r.Method == "GET" {
		err = streamAndCopyLogs(clientset, runName, w)
		if err != nil {
			return err
		}
	} else {
		// For the time being just show the results on stdout
		// TODO: Send them to the user with an http POST
		decoder := json.NewDecoder(r.Body)
		var l logMessage
		err := decoder.Decode(&l)
		if err != nil {
			return err
		}
		log.Printf("log %s %q", l.Stream, l.Log)
	}
	return nil
}

func delete(w http.ResponseWriter, r *http.Request) error {
	runName := mux.Vars(r)["id"]

	clientset, err := getClientSet()
	if err != nil {
		return err
	}

	err = deleteJob(clientset, runName)
	if err != nil {
		return err
	}
	err = deleteSecret(clientset, runName)
	if err != nil {
		return err
	}
	err = deleteFromStore(runName, "app", "tgz")
	if err != nil {
		return err
	}
	err = deleteFromStore(runName, "output", "")
	if err != nil {
		return err
	}
	err = deleteFromStore(runName, "exit-data", "json")
	if err != nil {
		return err
	}
	return deleteFromStore(runName, "cache", "tgz")
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
	authenticatedRouter.Handle("/app", appHandler(app)).Methods("PUT", "GET")
	authenticatedRouter.Handle("/cache", appHandler(cache)).Methods("PUT", "GET")
	authenticatedRouter.Handle("/output", appHandler(output)).Methods("PUT", "GET")
	authenticatedRouter.Handle("/exit-data", appHandler(exitData)).Methods("PUT", "GET")
	authenticatedRouter.Handle("/start", appHandler(start)).Methods("POST")
	authenticatedRouter.Handle("/logs", appHandler(logs)).Methods("POST", "GET")
	authenticatedRouter.Handle("", appHandler(delete)).Methods("DELETE")
	authenticatedRouter.Use(authenticate)
	router.Use(logRequests)

	log.Fatal(http.ListenAndServe(":8080", router))
}
