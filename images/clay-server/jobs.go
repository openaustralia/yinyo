package main

import (
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func createJob(clientset *kubernetes.Clientset, runName string, runOutput string) error {
	jobsClient := clientset.BatchV1().Jobs("clay")

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: runName,
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
							Name:    runName,
							Image:   "openaustralia/clay-scraper:v1",
							Command: []string{"/bin/run.sh", runName, runOutput},
							Env: []apiv1.EnvVar{
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

func deleteJob(clientset *kubernetes.Clientset, runName string) error {
	jobsClient := clientset.BatchV1().Jobs("clay")

	deletePolicy := metav1.DeletePropagationForeground
	err := jobsClient.Delete(runName, &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
	return err
}
