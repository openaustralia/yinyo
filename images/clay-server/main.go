package main

import (
	"fmt"
	"io"
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

func saveToStore(reader io.Reader, objectSize int64, scraperName string, fileName string, fileExtension string) error {
	minioClient, err := minio.New(
		// TODO: Get access key and password from secret
		"minio-service:9000", "admin", "changeme", false,
	)
	if err != nil {
		return err
	}

	path := fileName + "/" + scraperName
	if fileExtension != "" {
		path += "." + fileExtension
	}

	_, err = minioClient.PutObject(
		// TODO: Make bucket name configurable
		"clay",
		path,
		reader,
		objectSize,
		minio.PutObjectOptions{},
	)

	return err
}

func retrieveFromStore(scraperName string, fileName string, fileExtension string, writer io.Writer) error {
	minioClient, err := minio.New(
		// TODO: Get access key and password from secret
		"minio-service:9000", "admin", "changeme", false,
	)
	if err != nil {
		return err
	}

	path := fileName + "/" + scraperName
	if fileExtension != "" {
		path += "." + fileExtension
	}

	// TODO: Make bucket name configurable
	object, err := minioClient.GetObject("clay", path, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, object)
	return err
}

func deleteFromStore(scraperName string, fileName string, fileExtension string) error {
	minioClient, err := minio.New(
		// TODO: Get access key and password from secret
		"minio-service:9000", "admin", "changeme", false,
	)
	if err != nil {
		return err
	}

	path := fileName + "/" + scraperName
	if fileExtension != "" {
		path += "." + fileExtension
	}

	err = minioClient.RemoveObject("clay", path)
	return err
}

func createSecret(clientset *kubernetes.Clientset, scraperName string, runToken string) error {
	secretsClient := clientset.CoreV1().Secrets("default")
	secret := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: scraperName,
		},
		StringData: map[string]string{
			"run_token": runToken,
		},
	}
	_, err := secretsClient.Create(secret)
	return err
}

func deleteSecret(clientset *kubernetes.Clientset, scraperName string) error {
	secretsClient := clientset.CoreV1().Secrets("default")
	deletePolicy := metav1.DeletePropagationForeground
	err := secretsClient.Delete(scraperName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	return err
}

func createJob(clientset *kubernetes.Clientset, scraperName string, scraperOutput string) error {
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
							Image:   "openaustralia/clay-scraper:v1",
							Command: []string{"/bin/run.sh", scraperName, scraperOutput},
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
	_, err := jobsClient.Create(job)
	return err
}

func deleteJob(clientset *kubernetes.Clientset, scraperName string) error {
	jobsClient := clientset.BatchV1().Jobs("default")

	deletePolicy := metav1.DeletePropagationForeground
	err := jobsClient.Delete(scraperName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	return err
}

func create(w http.ResponseWriter, r *http.Request) {
	scraperName := mux.Vars(r)["id"]
	fmt.Println("create", scraperName)

	// Generate random token
	runToken := uniuri.NewLen(32)

	clientset, err := getClientSet()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = createSecret(clientset, scraperName, runToken)
	if err != nil {
		fmt.Println(err)
		return
	}
	// TODO: Return result as json
	fmt.Fprintln(w, runToken)
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
	scraperName := mux.Vars(r)["id"]
	runToken := r.Header.Get("Clay-Run-Token")

	fmt.Println(r.Method, fileName, scraperName)

	clientset, err := getClientSet()
	if err != nil {
		fmt.Println(err)
		return
	}

	actualRunToken, err := actualRunToken(clientset, scraperName)
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
		err = retrieveFromStore(scraperName, fileName, fileExtension, w)
	} else {
		err = saveToStore(r.Body, r.ContentLength, scraperName, fileName, fileExtension)
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

func actualRunToken(clientset *kubernetes.Clientset, scraperName string) (string, error) {
	// First get the actual run token from the secret
	secretsClient := clientset.CoreV1().Secrets("default")
	secret, err := secretsClient.Get(scraperName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	actualRunToken := string(secret.Data["run_token"])
	return actualRunToken, nil
}

func run(w http.ResponseWriter, r *http.Request) {
	scraperName := mux.Vars(r)["id"]
	scraperOutput := r.Header.Get("Clay-Scraper-Output")
	runToken := r.Header.Get("Clay-Run-Token")

	fmt.Println("run", scraperName)

	clientset, err := getClientSet()
	if err != nil {
		fmt.Println(err)
		return
	}

	actualRunToken, err := actualRunToken(clientset, scraperName)
	if err != nil {
		fmt.Println(err)
		return
	}

	if runToken != actualRunToken {
		// TODO: proper error with error code
		fmt.Println("Invalid run token")
		return
	}

	err = createJob(clientset, scraperName, scraperOutput)
	if err != nil {
		// TODO: Return error message to client
		// TODO: Remove secret
		fmt.Println(err)
		return
	}
}

func cleanup(w http.ResponseWriter, r *http.Request) {
	scraperName := mux.Vars(r)["id"]
	runToken := r.Header.Get("Clay-Run-Token")

	fmt.Println("run", scraperName)

	clientset, err := getClientSet()
	if err != nil {
		fmt.Println(err)
		return
	}

	actualRunToken, err := actualRunToken(clientset, scraperName)
	if err != nil {
		fmt.Println(err)
		return
	}

	if runToken != actualRunToken {
		// TODO: proper error with error code
		fmt.Println("Invalid run token")
		return
	}

	err = deleteJob(clientset, scraperName)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = deleteSecret(clientset, scraperName)
	if err != nil {
		fmt.Println(err)
		return
	}
	deleteFromStore(scraperName, "app", "tgz")
	deleteFromStore(scraperName, "output", "")
	deleteFromStore(scraperName, "cache", "tgz")
}

func whoAmI(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from Clay!")
}

func main() {
	fmt.Println("Clay is ready and waiting.")
	router := mux.NewRouter().StrictSlash(true)

	// TODO: Make url structure more RESTfull

	router.HandleFunc("/", whoAmI)
	router.HandleFunc("/scrapers/{id}/create", create).Methods("POST")
	router.HandleFunc("/scrapers/{id}/app", app).Methods("POST", "GET")
	router.HandleFunc("/scrapers/{id}/cache", cache).Methods("POST", "GET")
	router.HandleFunc("/scrapers/{id}/output", output).Methods("POST", "GET")
	router.HandleFunc("/scrapers/{id}/run", run).Methods("POST")
	router.HandleFunc("/scrapers/{id}/cleanup", cleanup).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}
