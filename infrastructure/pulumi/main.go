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
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"path/filepath"
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
	name         string
	appName      pulumi.String
	appPort      pulumi.Int
	outsidePort  pulumi.Int
	appLabels    pulumi.StringMap
	envs         map[string]string
	sharedVolume bool
}

func NewService(appName string, appPort int, outsidePort int, useSharedVolume bool) *Service {
	appLabels := pulumi.StringMap{
		"app": pulumi.String(appName),
	}
	return &Service{appName, pulumi.String(appName), pulumi.Int(appPort),
		pulumi.Int(outsidePort), appLabels, make(map[string]string, 0), useSharedVolume}
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

	// Ugly but i dont know how to change/transform/patch existing DeploymentArgs
	var volumeAnnotationValue pulumi.String
	var volumeArray corev1.VolumeArray
	var volumeMounts corev1.VolumeMountArray
	if service.sharedVolume {
		volumeAnnotationValue = "shared-volume:/tmp"
		volumeArray = corev1.VolumeArray{
			&corev1.VolumeArgs{
				Name: pulumi.String("shared-volume"),
				HostPath: &corev1.HostPathVolumeSourceArgs{
					Path: pulumi.String("/tmp"),
					Type: pulumi.String("DirectoryOrCreate"),
				},
			},
		}
		volumeMounts = corev1.VolumeMountArray{
			&corev1.VolumeMountArgs{
				MountPath: pulumi.String("/tmp"),
				Name:      pulumi.String("shared-volume"),
			},
		}
	}

	return &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Labels: service.appLabels,
			//Name:   service.appName,
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
						"dapr.io/volume-mounts-rw":                     volumeAnnotationValue,
					},
				},
				Spec: &corev1.PodSpecArgs{
					ImagePullSecrets: corev1.LocalObjectReferenceArray{
						&corev1.LocalObjectReferenceArgs{
							Name: pulumi.String("container-registry"),
						},
					},
					Volumes: volumeArray,
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
							VolumeMounts:    volumeMounts,
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
			//Name:   service.appName,
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
	mediaService := NewService("media", 8083, 8083, true)
	mediaService.AddContainerEnv("DAPR_PUBSUB_NAME", "pubsub")
	mediaService.AddContainerEnv("DAPR_TOPIC_NAME", "media")
	mediaService.AddContainerEnv("DAPR_S3_BINDING_NAME", "s3")
	services = append(services, mediaService)

	timelineService := NewService("timeline", 8082, 8082, false)
	timelineService.AddContainerEnv("DAPR_PUBSUB_NAME", "pubsub")
	timelineService.AddContainerEnv("DAPR_TOPIC_NAME", "status")
	timelineService.AddContainerEnv("DAPR_STATE_STORE_NAME", "statestore")
	services = append(services, timelineService)

	statusService := NewService("status", 8081, 8081, false)
	statusService.AddContainerEnv("DAPR_PUBSUB_NAME", "pubsub")
	statusService.AddContainerEnv("DAPR_TOPIC_NAME", "status")
	statusService.AddContainerEnv("DAPR_STATE_STORE_NAME", "statestore")
	services = append(services, statusService)

	userService := NewService("user", 8080, 8080, false)
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

		daprNamespace, err := corev1.NewNamespace(ctx, "dapr-system", &corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{Name: pulumi.String("dapr-system")},
		})
		if err != nil {
			return err
		}

		dapr, err := helm.NewChart(ctx, "dapr", helm.ChartArgs{
			Version:   pulumi.String("1.10"),
			Chart:     pulumi.String("dapr"),
			Namespace: daprNamespace.Metadata.Elem().Name().Elem(),
			FetchArgs: &helm.FetchArgs{
				Repo: pulumi.String("https://dapr.github.io/helm-charts/"),
			},
		}, pulumi.DependsOn([]pulumi.Resource{daprNamespace}))
		if err != nil {
			return err
		}

		minioNamespace, err := corev1.NewNamespace(ctx, "minio", &corev1.NamespaceArgs{
			Metadata: &metav1.ObjectMetaArgs{Name: pulumi.String("minio")},
		})
		if err != nil {
			return err
		}

		minio, err := helm.NewChart(ctx, "minio", helm.ChartArgs{
			Version:   pulumi.String("12.2.4"),
			Chart:     pulumi.String("minio"),
			Namespace: minioNamespace.Metadata.Elem().Name().Elem(),
			FetchArgs: &helm.FetchArgs{
				Repo: pulumi.String("https://charts.bitnami.com/bitnami"),
			},
			Values: pulumi.Map{
				"auth": pulumi.Map{
					"rootUser":     pulumi.String("minioadmin"),
					"rootPassword": pulumi.String("minioadmin"),
				},
				"defaultBuckets": pulumi.String("testbucket"),
			},
		}, pulumi.DependsOn([]pulumi.Resource{minioNamespace}))
		if err != nil {
			return err
		}

		redis, err := helm.NewChart(ctx, "redis", helm.ChartArgs{
			Version: pulumi.String("17.9.4"),
			Chart:   pulumi.String("redis"),
			FetchArgs: &helm.FetchArgs{
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

		daprComponents, err := yaml.NewConfigGroup(ctx, "dapr-components", &yaml.ConfigGroupArgs{
			Files: []string{filepath.Join("../../components/k8s", "*.yaml")},
		}, pulumi.DependsOnInputs(dapr.Ready))
		if err != nil {
			return err
		}

		for _, service := range services {
			_, err := NewServiceDeployment(ctx, service, pulumi.DependsOn([]pulumi.Resource{daprComponents}),
				pulumi.DependsOnInputs(dapr.Ready), pulumi.DependsOnInputs(minio.Ready), pulumi.DependsOnInputs(redis.Ready))
			if err != nil {
				return err
			}
		}
		return nil
	})
}
