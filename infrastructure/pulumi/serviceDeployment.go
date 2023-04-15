package main

import (
	"fmt"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
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

	serviceDeployment.Deployment, err = appsv1.NewDeployment(ctx, service.name, service.GetDeploymentArgs(), pulumi.Parent(serviceDeployment), pulumi.ReplaceOnChanges([]string{"*"}), pulumi.DeleteBeforeReplace(true))
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
	command      pulumi.StringArray
	args         pulumi.StringArray
	nodePort     pulumi.Int
}

func NewService(appName string, appPort int, outsidePort int, useSharedVolume bool) *Service {
	appLabels := pulumi.StringMap{
		"app": pulumi.String(appName),
	}
	return &Service{name: appName,
		appName:      pulumi.String(appName),
		appPort:      pulumi.Int(appPort),
		outsidePort:  pulumi.Int(outsidePort),
		appLabels:    appLabels,
		envs:         make(map[string]string, 0),
		sharedVolume: useSharedVolume}
}

func (service *Service) AddContainerCommands(cmds ...string) {
	if service.command == nil {
		service.command = make([]pulumi.StringInput, 0)
	}

	for _, cmd := range cmds {
		service.command = append(service.command, pulumi.String(cmd))
	}
}

func (service *Service) AddContainerArgs(args ...string) {
	if service.args == nil {
		service.args = make([]pulumi.StringInput, 0)
	}

	for _, arg := range args {
		service.args = append(service.args, pulumi.String(arg))
	}
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
						"dapr.io/sidecar-liveness-probe-delay-seconds": pulumi.String("15"),
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
							Command:         service.command,
							Args:            service.args,
						},
					},
				},
			},
		},
	}
}

func (service *Service) GetServiceArgs() *corev1.ServiceArgs {
	var servicePort *corev1.ServicePortArgs
	if service.nodePort != 0 {
		servicePort = &corev1.ServicePortArgs{
			NodePort:   service.nodePort,
			Port:       service.outsidePort,
			TargetPort: service.appPort,
		}
	} else {
		servicePort = &corev1.ServicePortArgs{
			Port:       service.outsidePort,
			TargetPort: service.appPort,
		}
	}

	return &corev1.ServiceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Labels: service.appLabels,
			Name:   service.appName,
		},
		Spec: &corev1.ServiceSpecArgs{
			Selector: pulumi.StringMap{
				"app": service.appName,
			},
			Ports: corev1.ServicePortArray{
				servicePort,
			},
			Type: pulumi.String("NodePort"),
		},
	}
}
