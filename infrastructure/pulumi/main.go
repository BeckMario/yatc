package main

import (
	"fmt"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"strconv"
)

type Service struct {
	appName string
	appPort int
	envs    map[string]string
}

func NewService(appName string, appPort int) *Service {
	return &Service{appName, appPort, make(map[string]string, 0)}
}

func (service *Service) AddEnv(key string, value string) {
	service.envs[key] = value
}

func (service *Service) GetDeployment() *appsv1.DeploymentArgs {
	appLabels := pulumi.StringMap{
		"app": pulumi.String(service.appName),
	}

	envArray := make([]corev1.EnvVarInput, 0)
	envArray = append(envArray, &corev1.EnvVarArgs{
		Name:  pulumi.String("PORT"),
		Value: pulumi.String(strconv.Itoa(service.appPort)),
	})

	for key, value := range service.envs {
		envArray = append(envArray, &corev1.EnvVarArgs{
			Name:  pulumi.String(key),
			Value: pulumi.String(value),
		})
	}
	return &appsv1.DeploymentArgs{
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: appLabels,
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: appLabels,
					Annotations: pulumi.StringMap{
						"dapr.io/enabled":            pulumi.String("true"),
						"dapr.io/app-id":             pulumi.String(service.appName),
						"dapr.io/app-port":           pulumi.String(strconv.Itoa(service.appPort)),
						"dapr.io/enable-api-logging": pulumi.String("true"),
					},
				},
				Spec: &corev1.PodSpecArgs{
					ImagePullSecrets: corev1.LocalObjectReferenceArray{
						&corev1.LocalObjectReferenceArgs{
							Name: pulumi.String("container-registry"),
						},
					},
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String(service.appName),
							Image: pulumi.String(fmt.Sprintf("reg.technicalonions.de/%s-service:latest", service.appName)),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(service.appPort),
								},
							},
							ImagePullPolicy: pulumi.String("Always"),
							Env:             corev1.EnvVarArray(envArray),
						},
					},
				},
			},
		},
	}
}

func main() {

	pulumi.Run(func(ctx *pulumi.Context) error {
		appName := "media"
		appLabels := pulumi.StringMap{
			"app": pulumi.String(appName),
		}
		appPort := 8080
		outsidePort := 8080

		_, err := corev1.NewService(ctx, appName, &corev1.ServiceArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Labels: appLabels,
			},
			Spec: &corev1.ServiceSpecArgs{
				Selector: pulumi.StringMap{
					"app": pulumi.String(appName),
				},
				Ports: corev1.ServicePortArray{
					&corev1.ServicePortArgs{
						Port:       pulumi.Int(outsidePort),
						TargetPort: pulumi.Any(appPort),
					},
				},
				Type: pulumi.String("LoadBalancer"),
			},
		})
		if err != nil {
			return err
		}

		_, err = appsv1.NewDeployment(ctx, appName, &appsv1.DeploymentArgs{
			Spec: &appsv1.DeploymentSpecArgs{
				Replicas: pulumi.Int(1),
				Selector: &metav1.LabelSelectorArgs{
					MatchLabels: appLabels,
				},
				Template: &corev1.PodTemplateSpecArgs{
					Metadata: &metav1.ObjectMetaArgs{
						Labels: appLabels,
						Annotations: pulumi.StringMap{
							"dapr.io/enabled":            pulumi.String("true"),
							"dapr.io/app-id":             pulumi.String(appName),
							"dapr.io/app-port":           pulumi.String(strconv.Itoa(appPort)),
							"dapr.io/enable-api-logging": pulumi.String("true"),
							"dapr.io/volume-mounts-rw":   pulumi.String("test-volume:/tmp"),
						},
					},
					Spec: &corev1.PodSpecArgs{
						Volumes: corev1.VolumeArray{
							&corev1.VolumeArgs{
								Name: pulumi.String("test-volume"),
								HostPath: &corev1.HostPathVolumeSourceArgs{
									Path: pulumi.String("/tmp"),
									Type: pulumi.String("DirectoryOrCreate"),
								},
							},
						},
						ImagePullSecrets: corev1.LocalObjectReferenceArray{
							&corev1.LocalObjectReferenceArgs{
								Name: pulumi.String("container-registry"),
							},
						},
						Containers: corev1.ContainerArray{
							&corev1.ContainerArgs{
								Name:  pulumi.String(appName),
								Image: pulumi.String("reg.technicalonions.de/media-service:latest"),
								Env: corev1.EnvVarArray{
									&corev1.EnvVarArgs{
										Name:  pulumi.String("PORT"),
										Value: pulumi.String(strconv.Itoa(appPort)),
									},
									&corev1.EnvVarArgs{
										Name:  pulumi.String("DAPR_PUBSUB_NAME"),
										Value: pulumi.String("pubsub"),
									},
									&corev1.EnvVarArgs{
										Name:  pulumi.String("DAPR_TOPIC_NAME"),
										Value: pulumi.String("media"),
									},
									&corev1.EnvVarArgs{
										Name:  pulumi.String("DAPR_S3_BINDING_NAME"),
										Value: pulumi.String("s3"),
									},
								},
								Ports: corev1.ContainerPortArray{
									&corev1.ContainerPortArgs{
										ContainerPort: pulumi.Int(8080),
									},
								},
								ImagePullPolicy: pulumi.String("Always"),
								VolumeMounts: corev1.VolumeMountArray{
									&corev1.VolumeMountArgs{
										MountPath: pulumi.String("/tmp"),
										Name:      pulumi.String("test-volume"),
									},
								},
							},
						},
					},
				},
			},
		})
		if err != nil {
			return err
		}
		return nil
	})
}
