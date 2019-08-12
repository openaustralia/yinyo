package main

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// The body of the request should contain the tarred & gzipped code
// to be run
// func run(w http.ResponseWriter, r *http.Request) {
// 	// scraperID := mux.Vars(r)["id"]
// 	// outputFilename := r.Header.Get("Clay-Scraper-Output")
//
// 	// For testing purposes let's just list the names of all the pods
// 	// currently running
// 	config, err := rest.InClusterConfig()
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
// 	if err != nil {
// 		panic(err.Error())
// 	}
// 	fmt.Fprintf(w, "There are %d pods in the cluster\n", len(pods.Items))
// 	// fmt.Fprintf(w, "scraperID: %s, outputFilename: %s", scraperID, outputFilename)
// }

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	for {
		pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
		time.Sleep(10 * time.Second)
	}

	// router := mux.NewRouter().StrictSlash(true)
	// router.HandleFunc("/scrapers/{id}/run", run).Methods("PUT")
	// log.Fatal(http.ListenAndServe(":8080", router))
}
