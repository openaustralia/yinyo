package jobdispatcher

import (
	"regexp"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type kubernetesClient struct {
	clientset *kubernetes.Clientset
}

func Kubernetes(clientset *kubernetes.Clientset) Client {
	return &kubernetesClient{clientset: clientset}
}

func (client *kubernetesClient) CreateJob(namePrefix string, runToken string) (string, error) {
	// We need to convert the user-supplied namePrefix to something that will
	// work in k8s. That means only alpha numeric characters and "-".
	// For instance no "/".

	// Matches any non alphanumeric character
	re := regexp.MustCompile(`[^[:alnum:]]`)
	convertedNamePrefix := re.ReplaceAllString(namePrefix, "-")

	secretsClient := client.clientset.CoreV1().Secrets("clay-scrapers")
	secret := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: convertedNamePrefix + "-",
		},
		StringData: map[string]string{
			"run_token": runToken,
		},
	}
	created, err := secretsClient.Create(secret)

	return created.ObjectMeta.Name, err
}
