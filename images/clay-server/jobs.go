package main

import (
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func createJob(clientset *kubernetes.Clientset, runName string, dockerImage string, command []string, env map[string]string) error {
	jobsClient := clientset.BatchV1().Jobs("clay-scrapers")

	autoMountServiceAccountToken := false
	backOffLimit := int32(0)
	// Let this run for a maximum of 24 hours
	activeDeadlineSeconds := int64(86400)

	environment := []apiv1.EnvVar{
		{
			Name: "CLAY_RUN_TOKEN",
			ValueFrom: &apiv1.EnvVarSource{
				SecretKeyRef: &apiv1.SecretKeySelector{
					LocalObjectReference: apiv1.LocalObjectReference{
						Name: runName,
					},
					Key: "run_token",
				},
			},
		},
	}
	// TODO: Check that runOptions.Env isn't trying to set CLAY_RUN_TOKEN
	// and warn the user if that is the case because the scraper will mysteriously not work
	for k, v := range env {
		environment = append(environment, apiv1.EnvVar{Name: k, Value: v})
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: runName,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backOffLimit,
			// Let this run for a maximum of 24 hours
			ActiveDeadlineSeconds: &activeDeadlineSeconds,
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					AutomountServiceAccountToken: &autoMountServiceAccountToken,
					RestartPolicy:                "Never",
					Containers: []apiv1.Container{
						{
							Name:    runName,
							Image:   dockerImage,
							Command: command,
							Env:     environment,
						},
					},
				},
			},
		},
	}
	_, err := jobsClient.Create(job)
	return err
}

func deleteJob(clientset *kubernetes.Clientset, runName string) error {
	jobsClient := clientset.BatchV1().Jobs("clay-scrapers")

	deletePolicy := metav1.DeletePropagationForeground
	err := jobsClient.Delete(runName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	if err != nil {
		// Don't error if it's just that the job doesn't exist
		if err.(*apierrors.StatusError).ErrStatus.Reason == metav1.StatusReasonNotFound {
			return nil
		}
		return err
	}
	return nil
}
