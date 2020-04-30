package jobdispatcher

import (
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type kubernetesClient struct {
	clientset *kubernetes.Clientset
}

const namespace = "yinyo-runs"

// NewKubernetes returns the Kubernetes implementation of Client
func NewKubernetes() (Jobs, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	k := &kubernetesClient{clientset: clientset}
	return k, nil
}

// maxRunTime is the maximum number of seconds that the job is allowed to take. If it exceeds this limit it will get stopped automatically
// memory is the amount of memory (in bytes) that is allocated to this job. If more is used it will get killed. Note that this
// memory is effectively reserved for a job so allocating too much if it's not used is wasteful.
func (client *kubernetesClient) Create(runID string, dockerImage string, command []string, maxRunTime int64, memory int64) error {
	jobsClient := client.clientset.BatchV1().Jobs(namespace)

	autoMountServiceAccountToken := false
	// Allow the job to get restarted up to 5 times before it's considered failed
	backOffLimit := int32(5)
	memoryQuantity := resource.NewQuantity(memory, resource.BinarySI)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: runID,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:          &backOffLimit,
			ActiveDeadlineSeconds: &maxRunTime,
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					AutomountServiceAccountToken: &autoMountServiceAccountToken,
					RestartPolicy:                "OnFailure",
					Containers: []apiv1.Container{
						{
							Name:    runID,
							Image:   dockerImage,
							Command: command,
							Resources: apiv1.ResourceRequirements{
								Requests: apiv1.ResourceList{
									apiv1.ResourceMemory: *memoryQuantity,
									apiv1.ResourceCPU:    resource.MustParse("250m"), // 1/4 of a vCPU
								},
								Limits: apiv1.ResourceList{
									apiv1.ResourceMemory: *memoryQuantity,
									apiv1.ResourceCPU:    resource.MustParse("1000m"), // One vCPU
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

func (client *kubernetesClient) Delete(runID string) error {
	jobsClient := client.clientset.BatchV1().Jobs(namespace)

	deletePolicy := metav1.DeletePropagationForeground
	err := jobsClient.Delete(runID, &metav1.DeleteOptions{
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
