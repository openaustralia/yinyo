package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dchest/uniuri"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v6"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }

// The body of the request should contain the tarred & gzipped code
// to be run
func run(w http.ResponseWriter, r *http.Request) {
	scraperName := mux.Vars(r)["id"]
	scraperOutput := r.Header.Get("Clay-Scraper-Output")

	minioClient, err := minio.New(
		// TODO: Get access key and password from secret
		"minio-service:9000", "admin", "changeme", false,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = minioClient.PutObject(
		"clay",
		"app/"+scraperName+".tgz",
		r.Body,
		r.ContentLength,
		minio.PutObjectOptions{},
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Generate random token
	runToken := uniuri.NewLen(32)

	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
		return
	}

	secretsClient := clientset.CoreV1().Secrets("default")
	secret := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: scraperName,
		},
		StringData: map[string]string{
			"run_token": runToken,
		},
	}
	_, err = secretsClient.Create(secret)
	if err != nil {
		fmt.Println(err)
		return
	}

	jobsClient := clientset.BatchV1().Jobs("default")

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: scraperName,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: int32Ptr(0),
			// Let this run for a maximum of 24 hours
			ActiveDeadlineSeconds: int64Ptr(86400),
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					RestartPolicy: "Never",
					Containers: []apiv1.Container{
						{
							Name:    scraperName,
							Image:   "clay-scraper",
							Command: []string{"/bin/run.sh", scraperName, scraperOutput},
							// Doing this so that we use the local image while we're developing
							ImagePullPolicy: "Never",
							Env: []apiv1.EnvVar{
								{
									Name: "CLAY_RUN_TOKEN",
									ValueFrom: &apiv1.EnvVarSource{
										SecretKeyRef: &apiv1.SecretKeySelector{
											LocalObjectReference: apiv1.LocalObjectReference{
												Name: scraperName,
											},
											Key: "run_token",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	_, err = jobsClient.Create(job)
	if err != nil {
		// TODO: Return error message to client
		fmt.Println(err)
		return
	}

	// TODO: Return result as json
	fmt.Fprintln(w, runToken)
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/scrapers/{id}/run", run).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}
