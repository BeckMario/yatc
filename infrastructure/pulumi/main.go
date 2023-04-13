package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"path/filepath"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"strconv"
)

type ServiceDeployment struct {
	pulumi.ResourceState

	Deployment *appsv1.Deployment
	Service    *corev1.Service
}

func NewServiceDeployment(ctx *pulumi.Context, service *Service, opts ...pulumi.ResourceOption) (*ServiceDeployment, error) {
	serviceDeployment := &ServiceDeployment{}

	err := ctx.RegisterComponentResource("yatc:component:ServiceDeployment", service.name, serviceDeployment, opts...)
	if err != nil {
		return nil, err
	}

	serviceDeployment.Deployment, err = appsv1.NewDeployment(ctx, service.name, service.GetDeploymentArgs(), pulumi.Parent(serviceDeployment))
	if err != nil {
		return nil, err
	}

	serviceDeployment.Service, err = corev1.NewService(ctx, service.name, service.GetServiceArgs(), pulumi.Parent(serviceDeployment))
	if err != nil {
		return nil, err
	}

	return serviceDeployment, nil
}

type Service struct {
	name        string
	appName     pulumi.String
	appPort     pulumi.Int
	outsidePort pulumi.Int
	appLabels   pulumi.StringMap
	envs        map[string]string
}

func NewService(appName string, appPort int, outsidePort int) *Service {
	appLabels := pulumi.StringMap{
		"app": pulumi.String(appName),
	}
	return &Service{appName, pulumi.String(appName), pulumi.Int(appPort),
		pulumi.Int(outsidePort), appLabels, make(map[string]string, 0)}
}

func (service *Service) AddContainerEnv(key string, value string) {
	service.envs[key] = value
}

func (service *Service) getEnvVarArray() corev1.EnvVarArray {
	portStringOutput := service.appPort.ToIntOutput().ApplyT(func(port int) string {
		return strconv.Itoa(port)
	}).(pulumi.StringOutput)

	envArray := make([]corev1.EnvVarInput, 0)
	envArray = append(envArray, &corev1.EnvVarArgs{
		Name:  pulumi.String("PORT"),
		Value: portStringOutput,
	})

	for key, value := range service.envs {
		envArray = append(envArray, &corev1.EnvVarArgs{
			Name:  pulumi.String(key),
			Value: pulumi.String(value),
		})
	}

	return envArray
}

func (service *Service) GetDeploymentArgs() *appsv1.DeploymentArgs {
	portStringOutput := service.appPort.ToIntOutput().ApplyT(func(port int) string {
		return strconv.Itoa(port)
	}).(pulumi.StringOutput)

	return &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Labels: service.appLabels,
			Name:   service.appName,
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: service.appLabels,
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: service.appLabels,
					Annotations: pulumi.StringMap{
						"dapr.io/enabled":                              pulumi.String("true"),
						"dapr.io/app-id":                               pulumi.Sprintf("%s-service", service.name),
						"dapr.io/app-port":                             portStringOutput,
						"dapr.io/enable-api-logging":                   pulumi.String("true"),
						"dapr.io/log-level":                            pulumi.String("debug"),
						"dapr.io/sidecar-liveness-probe-delay-seconds": pulumi.String("20"),
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
							Name:  service.appName,
							Image: pulumi.String(fmt.Sprintf("reg.technicalonions.de/%s-service:latest", service.appName)),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: service.appPort,
								},
							},
							ImagePullPolicy: pulumi.String("Always"),
							Env:             service.getEnvVarArray(),
						},
					},
				},
			},
		},
	}
}

func (service *Service) GetServiceArgs() *corev1.ServiceArgs {
	return &corev1.ServiceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Labels: service.appLabels,
		},
		Spec: &corev1.ServiceSpecArgs{
			Selector: pulumi.StringMap{
				"app": service.appName,
			},
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       service.outsidePort,
					TargetPort: service.appPort,
				},
			},
			Type: pulumi.String("NodePort"),
		},
	}
}

type DockerConfigAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Auth     string `json:"auth"`
}

type DockerConfig struct {
	Auths map[string]DockerConfigAuth `json:"auths"`
}

func main() {

	services := make([]*Service, 0)
	mediaService := NewService("media", 8083, 8083)
	mediaService.AddContainerEnv("DAPR_PUBSUB_NAME", "pubsub")
	mediaService.AddContainerEnv("DAPR_TOPIC_NAME", "media")
	mediaService.AddContainerEnv("DAPR_TOPIC_NAME", "s3")
	services = append(services, mediaService)

	timelineService := NewService("timeline", 8082, 8082)
	timelineService.AddContainerEnv("DAPR_PUBSUB_NAME", "pubsub")
	timelineService.AddContainerEnv("DAPR_TOPIC_NAME", "status")
	timelineService.AddContainerEnv("DAPR_STATE_STORE_NAME", "statestore")
	services = append(services, timelineService)

	statusService := NewService("status", 8081, 8081)
	statusService.AddContainerEnv("DAPR_PUBSUB_NAME", "pubsub")
	statusService.AddContainerEnv("DAPR_TOPIC_NAME", "status")
	statusService.AddContainerEnv("DATABASE", "TODO")
	services = append(services, statusService)

	userService := NewService("user", 8080, 8080)
	userService.AddContainerEnv("DAPR_STATE_STORE_NAME", "statestore")
	services = append(services, userService)

	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		registry := cfg.Require("docker-registry")
		username := cfg.Require("docker-username")
		password := cfg.RequireSecret("docker-password")

		imagePullSecret := &corev1.SecretArgs{
			Data: &pulumi.StringMap{
				".dockerconfigjson": password.ApplyT(func(password string) (string, error) {
					authB64 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))

					dockerConfig := DockerConfig{Auths: map[string]DockerConfigAuth{registry: {
						Username: username,
						Password: password,
						Auth:     authB64,
					}}}

					jsonBytes, err := json.Marshal(dockerConfig)
					if err != nil {
						return "", err
					}
					return base64.StdEncoding.EncodeToString(jsonBytes), nil
				}).(pulumi.StringOutput),
			},
			Metadata: &metav1.ObjectMetaArgs{
				Name: pulumi.String("container-registry"),
			},
			Type: pulumi.String("kubernetes.io/dockerconfigjson"),
		}

		_, err := corev1.NewSecret(ctx, "container-registry", imagePullSecret)
		if err != nil {
			return err
		}

		dapr, err := helm.NewRelease(ctx, "dapr", &helm.ReleaseArgs{
			Version:         pulumi.String("1.10"),
			Chart:           pulumi.String("dapr"),
			Namespace:       pulumi.String("dapr-system"),
			CreateNamespace: pulumi.Bool(true),
			RepositoryOpts: &helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://dapr.github.io/helm-charts/"),
			},
		})
		if err != nil {
			return err
		}

		redis, err := helm.NewRelease(ctx, "redis", &helm.ReleaseArgs{
			Version: pulumi.String("17.9.4"),
			Chart:   pulumi.String("redis"),
			RepositoryOpts: &helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://charts.bitnami.com/bitnami"),
			},
			Values: pulumi.Map{
				"auth": pulumi.Map{
					"password": pulumi.String("redis"),
				},
			},
		})
		if err != nil {
			return err
		}

		minio, err := helm.NewRelease(ctx, "minio", &helm.ReleaseArgs{
			Version:         pulumi.String("12.2.4"),
			Chart:           pulumi.String("minio"),
			Namespace:       pulumi.String("minio"),
			CreateNamespace: pulumi.Bool(true),
			RepositoryOpts: &helm.RepositoryOptsArgs{
				Repo: pulumi.String("https://charts.bitnami.com/bitnami"),
			},
			Values: pulumi.Map{
				"auth": pulumi.Map{
					"rootUser":     pulumi.String("minioadmin"),
					"rootPassword": pulumi.String("minioadmin"),
				},
				"defaultBuckets": pulumi.String("testbucket"),
			},
		})
		if err != nil {
			return err
		}

		daprComponents, err := yaml.NewConfigGroup(ctx, "dapr-components", &yaml.ConfigGroupArgs{
			Files: []string{filepath.Join("../../components/k8s", "*.yaml")},
		}, pulumi.DependsOn([]pulumi.Resource{dapr}))
		if err != nil {
			return err
		}

		for _, service := range services {
			_, err := NewServiceDeployment(ctx, service, pulumi.DependsOn([]pulumi.Resource{dapr, redis, minio, daprComponents}))
			if err != nil {
				return err
			}
		}
		return nil
	})

	/*pulumi.Run(func(ctx *pulumi.Context) error {
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
			Metadata: &metav1.ObjectMetaArgs{
				Labels: appLabels,
				Name:   pulumi.String(appName),
			},
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
	})*/
}