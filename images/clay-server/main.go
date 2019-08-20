package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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

func create(w http.ResponseWriter, r *http.Request) {
	// TODO: Make the scraper_name optional
	// TODO: Do we make sure that there is only one scraper_name used?
	scraperName := r.URL.Query()["scraper_name"][0]

	clientset, err := getClientSet()
	if err != nil {
		fmt.Println(err)
		return
	}

	runName, runToken, err := createSecret(clientset, scraperName)
	if err != nil {
		fmt.Println(err)
		return
	}

	createResult := createResult{
		RunName:  runName,
		RunToken: runToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createResult)
}

func store(w http.ResponseWriter, r *http.Request, fileName string, fileExtension string) {
	runName := mux.Vars(r)["id"]

	if r.Method == "GET" {
		err := retrieveFromStore(runName, fileName, fileExtension, w)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		err := saveToStore(r.Body, r.ContentLength, runName, fileName, fileExtension)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

// The body of the request should contain the tarred & gzipped code
func app(w http.ResponseWriter, r *http.Request) {
	store(w, r, "app", "tgz")
}

// The body of the request should contain the tarred & gzipped cache
func cache(w http.ResponseWriter, r *http.Request) {
	store(w, r, "cache", "tgz")
}

func output(w http.ResponseWriter, r *http.Request) {
	store(w, r, "output", "")
}

func start(w http.ResponseWriter, r *http.Request) {
	runName := mux.Vars(r)["id"]
	runOutput := r.URL.Query()["output"][0]

	clientset, err := getClientSet()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = createJob(clientset, runName, runOutput)
	if err != nil {
		// TODO: Return error message to client
		// TODO: Remove secret
		fmt.Println(err)
		return
	}
}

type logMessage struct {
	// TODO: Make the stream an enum
	Log, Stream string
}

func logs(w http.ResponseWriter, r *http.Request) {
	runName := mux.Vars(r)["id"]

	clientset, err := getClientSet()
	if err != nil {
		fmt.Println(err)
		return
	}

	if r.Method == "GET" {
		err = streamAndCopyLogs(clientset, runName, w)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		// For the time being just show the results on stdout
		// TODO: Send them to the user with an http POST
		decoder := json.NewDecoder(r.Body)
		var l logMessage
		err := decoder.Decode(&l)
		if err != nil {
			fmt.Println(err)
			return
		}
		log.Printf("log %s %q", l.Stream, l.Log)
	}
}

func delete(w http.ResponseWriter, r *http.Request) {
	runName := mux.Vars(r)["id"]

	clientset, err := getClientSet()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = deleteJob(clientset, runName)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = deleteSecret(clientset, runName)
	if err != nil {
		fmt.Println(err)
		return
	}
	deleteFromStore(runName, "app", "tgz")
	deleteFromStore(runName, "output", "")
	deleteFromStore(runName, "cache", "tgz")
}

func whoAmI(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from Clay!")
}

// Middleware that logs the request uri
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

// Middleware function, which will be called for each request
func authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		runName := mux.Vars(r)["id"]
		runToken := r.Header.Get("Clay-Run-Token")

		clientset, err := getClientSet()
		if err != nil {
			log.Println(err)
			http.Error(w, "Could not contact kubernetes", http.StatusInternalServerError)
			return
		}

		actualRunToken, err := actualRunToken(clientset, runName)
		if err != nil {
			log.Println(err)
			http.Error(w, "Could not contact kubernetes", http.StatusInternalServerError)
			return
		}

		if runToken != actualRunToken {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("Clay is ready and waiting.")
	router := mux.NewRouter().StrictSlash(true)

	// TODO: Use more conventional basic auth
	router.HandleFunc("/", whoAmI)
	router.HandleFunc("/runs", create).Methods("POST")

	authenticatedRouter := router.PathPrefix("/runs/{id}").Subrouter()
	authenticatedRouter.HandleFunc("/app", app).Methods("PUT", "GET")
	authenticatedRouter.HandleFunc("/cache", cache).Methods("PUT", "GET")
	authenticatedRouter.HandleFunc("/output", output).Methods("PUT", "GET")
	authenticatedRouter.HandleFunc("/start", start).Methods("POST")
	authenticatedRouter.HandleFunc("/logs", logs).Methods("POST", "GET")
	authenticatedRouter.HandleFunc("", delete).Methods("DELETE")
	authenticatedRouter.Use(authenticate)
	router.Use(logRequests)

	log.Fatal(http.ListenAndServe(":8080", router))
}
