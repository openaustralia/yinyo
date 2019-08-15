package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dchest/uniuri"
	"github.com/gorilla/mux"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }

type createResult struct {
	RunName  string `json:"run_name"`
	RunToken string `json:"run_token"`
}

func create(w http.ResponseWriter, r *http.Request) {
	// TODO: Make the scraper_name optional
	// TODO: Do we make sure that there is only one scraper_name used?
	scraperName := r.URL.Query()["scraper_name"][0]

	fmt.Println("create", scraperName)

	// Generate random token
	runToken := uniuri.NewLen(32)

	clientset, err := getClientSet()
	if err != nil {
		fmt.Println(err)
		return
	}

	secret, err := createSecret(clientset, scraperName, runToken)
	if err != nil {
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	createResult := createResult{
		RunName:  secret.ObjectMeta.Name,
		RunToken: runToken,
	}
	json.NewEncoder(w).Encode(createResult)
}

func getClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	return clientset, err
}

func store(w http.ResponseWriter, r *http.Request, fileName string, fileExtension string) {
	runName := mux.Vars(r)["id"]
	runToken := r.Header.Get("Clay-Run-Token")

	fmt.Println(r.Method, fileName, runName)

	clientset, err := getClientSet()
	if err != nil {
		fmt.Println(err)
		return
	}

	actualRunToken, err := actualRunToken(clientset, runName)
	if err != nil {
		fmt.Println(err)
		return
	}

	if runToken != actualRunToken {
		// TODO: proper error with error code
		fmt.Println("Invalid run token")
		return
	}

	if r.Method == "GET" {
		err = retrieveFromStore(runName, fileName, fileExtension, w)
	} else {
		err = saveToStore(r.Body, r.ContentLength, runName, fileName, fileExtension)
	}
	if err != nil {
		fmt.Println(err)
		return
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
	scraperOutput := r.Header.Get("Clay-Scraper-Output")
	runToken := r.Header.Get("Clay-Run-Token")

	fmt.Println("start", runName)

	clientset, err := getClientSet()
	if err != nil {
		fmt.Println(err)
		return
	}

	actualRunToken, err := actualRunToken(clientset, runName)
	if err != nil {
		fmt.Println(err)
		return
	}

	if runToken != actualRunToken {
		// TODO: proper error with error code
		fmt.Println("Invalid run token")
		return
	}

	err = createJob(clientset, runName, scraperOutput)
	if err != nil {
		// TODO: Return error message to client
		// TODO: Remove secret
		fmt.Println(err)
		return
	}
}

// Return pod name
func waitForPodToStart(clientset *kubernetes.Clientset, runName string) (string, error) {
	podsClient := clientset.CoreV1().Pods("default")
	// TODO: Don't wait forever
	for {
		list, err := podsClient.List(metav1.ListOptions{
			LabelSelector: "job-name=" + runName,
		})
		if err != nil {
			return "", err
		}
		if len(list.Items) > 0 {
			podName := list.Items[0].ObjectMeta.Name
			// Now that we know the pod exists, let's check if it has started
			pod, err := podsClient.Get(podName, metav1.GetOptions{})
			if err != nil {
				return "", err
			}
			if pod.Status.Phase != apiv1.PodPending {
				return podName, nil
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func streamAndCopyLogs(clientset *kubernetes.Clientset, runName string, w http.ResponseWriter) error {
	podsClient := clientset.CoreV1().Pods("default")

	podName, err := waitForPodToStart(clientset, runName)
	if err != nil {
		return err
	}

	req := podsClient.GetLogs(podName, &apiv1.PodLogOptions{
		Follow: true,
	})
	podLogs, err := req.Stream()
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
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func logs(w http.ResponseWriter, r *http.Request) {
	runName := mux.Vars(r)["id"]
	runToken := r.Header.Get("Clay-Run-Token")

	fmt.Println("logs", runName)

	clientset, err := getClientSet()
	if err != nil {
		fmt.Println(err)
		return
	}

	actualRunToken, err := actualRunToken(clientset, runName)
	if err != nil {
		fmt.Println(err)
		return
	}

	if runToken != actualRunToken {
		// TODO: proper error with error code
		fmt.Println("Invalid run token")
		return
	}

	err = streamAndCopyLogs(clientset, runName, w)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func delete(w http.ResponseWriter, r *http.Request) {
	runName := mux.Vars(r)["id"]
	runToken := r.Header.Get("Clay-Run-Token")

	fmt.Println("delete", runName)

	clientset, err := getClientSet()
	if err != nil {
		fmt.Println(err)
		return
	}

	actualRunToken, err := actualRunToken(clientset, runName)
	if err != nil {
		fmt.Println(err)
		return
	}

	if runToken != actualRunToken {
		// TODO: proper error with error code
		fmt.Println("Invalid run token")
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

func main() {
	fmt.Println("Clay is ready and waiting.")
	router := mux.NewRouter().StrictSlash(true)

	// TODO: Use more conventional basic auth
	router.HandleFunc("/", whoAmI)
	router.HandleFunc("/runs", create).Methods("POST")
	router.HandleFunc("/runs/{id}/app", app).Methods("PUT", "GET")
	router.HandleFunc("/runs/{id}/cache", cache).Methods("PUT", "GET")
	router.HandleFunc("/runs/{id}/output", output).Methods("PUT", "GET")
	// TODO: Put scraper output as a parameter in the url
	router.HandleFunc("/runs/{id}/start", start).Methods("POST")
	router.HandleFunc("/runs/{id}/logs", logs).Methods("GET")
	router.HandleFunc("/runs/{id}", delete).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8080", router))
}
