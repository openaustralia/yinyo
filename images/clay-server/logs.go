package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Return pod name
func waitForPodToStart(clientset *kubernetes.Clientset, runName string) (string, error) {
	podsClient := clientset.CoreV1().Pods("clay")
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
	podsClient := clientset.CoreV1().Pods("clay")

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
