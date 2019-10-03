package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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
