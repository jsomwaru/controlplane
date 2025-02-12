package tfexec

import (
	"fmt"

	controlplaneiov1 "controlplane.io/terraformrequest/api/v1"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultTerraformImageRepo = "hashicorp/terraform"
	defaultImageTag           = "1.9.6"
)

func CreateTerraformApplyJob(tfreq *controlplaneiov1.TerraformRequest) (*kbatch.Job, *corev1.ConfigMap) {
	code := tfreq.Spec.Inline
	configMapName := fmt.Sprintf("%s-cm", tfreq.ObjectMeta.Name)
	jobName := fmt.Sprintf("tfapply-%s", tfreq.ObjectMeta.Name)
	annotations := make(map[string]string)
	annotations["controlplane.io/owner"] = tfreq.ObjectMeta.Name
	var backOffLimit int32 = 0
	cm := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        configMapName,
			Namespace:   tfreq.ObjectMeta.Namespace,
			Annotations: annotations,
		},
		Data: map[string]string{
			"code": code,
		},
	}
	command := fmt.Sprintf("cp /terraform/* /opt/work && cd /opt/work && terraform init && terraform %s", tfreq.Spec.Operation)
	job := kbatch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: tfreq.ObjectMeta.Namespace,
		},
		Spec: kbatch.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    jobName,
							Image:   fmt.Sprintf("%s:%s", defaultTerraformImageRepo, defaultImageTag),
							Command: []string{"/bin/sh"},
							Args:    []string{"-c", command},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "code",
									MountPath: "/terraform",
								},
								{
									Name:      "workdir",
									MountPath: "/opt/work",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "TF_DATA_DIR",
									Value: "/tmp",
								},
							},
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: tfreq.Spec.CredentialsSecretName,
										},
									},
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "code",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMapName,
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "code",
											Path: "main.tf",
										},
									},
								},
							},
						},
						{
							Name: "workdir",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					RestartPolicy: "Never",
				},
			},
			BackoffLimit: &backOffLimit,
		},
	}

	return &job, &cm
}
