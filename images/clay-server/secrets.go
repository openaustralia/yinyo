package main

import (
	"regexp"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// TODO: Generate the runToken in this method
func createSecret(clientset *kubernetes.Clientset, scraperName string, runToken string) (*apiv1.Secret, error) {
	// We need to convert the user-supplied scraperName to something that will
	// work in k8s. That means only alpha numeric characters and "-".
	// For instance no "/".

	// Matches any non alphanumeric character
	re := regexp.MustCompile(`[^[:alnum:]]`)
	convertedScraperName := re.ReplaceAllString(scraperName, "-")

	secretsClient := clientset.CoreV1().Secrets("default")
	secret := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: convertedScraperName + "-",
		},
		StringData: map[string]string{
			"run_token": runToken,
		},
	}
	created, err := secretsClient.Create(secret)
	return created, err
}

func deleteSecret(clientset *kubernetes.Clientset, runName string) error {
	secretsClient := clientset.CoreV1().Secrets("default")
	deletePolicy := metav1.DeletePropagationForeground
	err := secretsClient.Delete(runName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	return err
}

func actualRunToken(clientset *kubernetes.Clientset, runName string) (string, error) {
	// First get the actual run token from the secret
	secretsClient := clientset.CoreV1().Secrets("default")
	secret, err := secretsClient.Get(runName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	actualRunToken := string(secret.Data["run_token"])
	return actualRunToken, nil
}
