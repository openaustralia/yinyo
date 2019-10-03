package main

import (
	"regexp"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func createSecret(clientset *kubernetes.Clientset, namePrefix string, runToken string) (string, error) {
	// We need to convert the user-supplied namePrefix to something that will
	// work in k8s. That means only alpha numeric characters and "-".
	// For instance no "/".

	// Matches any non alphanumeric character
	re := regexp.MustCompile(`[^[:alnum:]]`)
	convertedNamePrefix := re.ReplaceAllString(namePrefix, "-")

	secretsClient := clientset.CoreV1().Secrets("clay-scrapers")
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

func deleteSecret(clientset *kubernetes.Clientset, runName string) error {
	secretsClient := clientset.CoreV1().Secrets("clay-scrapers")
	deletePolicy := metav1.DeletePropagationForeground
	err := secretsClient.Delete(runName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	return err
}

func actualRunToken(clientset *kubernetes.Clientset, runName string) (string, error) {
	// First get the actual run token from the secret
	secretsClient := clientset.CoreV1().Secrets("clay-scrapers")
	secret, err := secretsClient.Get(runName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	actualRunToken := string(secret.Data["run_token"])
	return actualRunToken, nil
}
