package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
	"path/filepath"
	openfunctionv1 "yatc/crds/kubernetes/core/v1beta1"
)

type DockerConfigAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Auth     string `json:"auth"`
}

type DockerConfig struct {
	Auths map[string]DockerConfigAuth `json:"auths"`
}

func main() {
	services := make([]Service, 0)
	mediaService := NewDaprService("media", 8083, 8083, true)
	mediaService.AddContainerEnv("DAPR_PUBSUB_NAME", "pubsub")
	mediaService.AddContainerEnv("DAPR_TOPIC_NAME", "media")
	mediaService.AddContainerEnv("DAPR_S3_BINDING_NAME", "s3")
	services = append(services, mediaService)

	timelineService := NewDaprService("timeline", 8082, 8082, false)
	timelineService.AddContainerEnv("DAPR_PUBSUB_NAME", "pubsub")
	timelineService.AddContainerEnv("DAPR_TOPIC_NAME", "status")
	timelineService.AddContainerEnv("DAPR_STATE_STORE_NAME", "statestore")
	services = append(services, timelineService)

	statusService := NewDaprService("status", 8081, 8081, false)
	statusService.AddContainerEnv("DAPR_PUBSUB_NAME", "pubsub")
	statusService.AddContainerEnv("DAPR_TOPIC_NAME", "status")
	statusService.AddContainerEnv("DAPR_STATE_STORE_NAME", "statestore")
	services = append(services, statusService)

	userService := NewDaprService("user", 8080, 8080, false)
	userService.AddContainerEnv("DAPR_STATE_STORE_NAME", "statestore")
	services = append(services, userService)

	krakendGateway := NewDaprService("krakend", 8080, 8080, false)
	krakendGateway.AddContainerCommands("/usr/bin/krakend")
	krakendGateway.AddContainerArgs("run", "-d", "-c", "/etc/krakend/krakend.json", "-p", "8080")
	krakendGateway.AddContainerEnv("KRAKEND_PORT", "8080")
	krakendGateway.nodePort = pulumi.Int(30442)
	services = append(services, krakendGateway)

	loginService := NewDaprService("login", 8084, 8084, false)
	services = append(services, loginService)

	zipkinService := NewService("zipkin", 9411, 9411, "openzipkin/zipkin")
	services = append(services, zipkinService)

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

		dapr, err := createDapr(ctx)
		if err != nil {
			return err
		}

		minio, err := createMinio(ctx)
		if err != nil {
			return err
		}

		redis, err := createRedis(ctx)
		if err != nil {
			return err
		}

		openfunction, err2 := createOpenFunction(ctx, dapr)
		if err2 != nil {
			return err2
		}

		daprComponents, err := yaml.NewConfigGroup(ctx, "dapr-components", &yaml.ConfigGroupArgs{
			Files: []string{filepath.Join("../dapr-components/k8s", "*.yaml")},
		}, pulumi.DependsOnInputs(dapr.Ready))
		if err != nil {
			return err
		}

		dependencies := []pulumi.ResourceOption{pulumi.DependsOn([]pulumi.Resource{daprComponents, openfunction}),
			pulumi.DependsOnInputs(dapr.Ready), pulumi.DependsOnInputs(minio.Ready), pulumi.DependsOnInputs(redis.Ready)}

		_, err = openfunctionv1.NewFunction(ctx, "media-conversion", &openfunctionv1.FunctionArgs{
			Metadata: metav1.ObjectMetaArgs{Name: pulumi.String("media-conversion")},
			Spec: openfunctionv1.FunctionSpecArgs{
				Image: pulumi.String("reg.technicalonions.de/media-conversion:latest"),
				ImageCredentials: openfunctionv1.FunctionSpecImagecredentialsArgs{
					Name: pulumi.String("container-registry")},
				Serving: openfunctionv1.FunctionSpecServingArgs{
					Annotations: pulumi.StringMap{
						"dapr.io/config": pulumi.String("tracing"),
					},
					Runtime: pulumi.String("async"),
					Inputs: openfunctionv1.FunctionSpecServingInputsArray{
						&openfunctionv1.FunctionSpecServingInputsArgs{
							Name:      pulumi.String("subscriber"),
							Component: pulumi.String("redis-server"),
							Topic:     pulumi.String("media"),
						},
					},
					Pubsub: &openfunctionv1.FunctionSpecServingPubsubMap{
						"redis-server": &openfunctionv1.FunctionSpecServingPubsubArgs{
							Metadata: &openfunctionv1.FunctionSpecServingPubsubMetadataArray{
								&openfunctionv1.FunctionSpecServingPubsubMetadataArgs{
									Name:  pulumi.String("redisHost"),
									Value: pulumi.String("redis-master:6379"),
								},
								&openfunctionv1.FunctionSpecServingPubsubMetadataArgs{
									Name:  pulumi.String("redisPassword"),
									Value: pulumi.String("redis"),
								},
								&openfunctionv1.FunctionSpecServingPubsubMetadataArgs{
									Name:  pulumi.String("consumerID"),
									Value: pulumi.String("subscriber"),
								},
								&openfunctionv1.FunctionSpecServingPubsubMetadataArgs{
									Name:  pulumi.String("enableTLS"),
									Value: pulumi.String("false"),
								},
							},
							Type:    pulumi.String("pubsub.redis"),
							Version: pulumi.String("v1"),
						},
					},
					Template: &openfunctionv1.FunctionSpecServingTemplateArgs{
						Containers: openfunctionv1.FunctionSpecServingTemplateContainersArray{
							&openfunctionv1.FunctionSpecServingTemplateContainersArgs{
								Name: pulumi.String("function"),
								Env: openfunctionv1.FunctionSpecServingTemplateContainersEnvArray{
									&openfunctionv1.FunctionSpecServingTemplateContainersEnvArgs{
										Name:  pulumi.String("S3_ENDPOINT"),
										Value: pulumi.String("minio.minio.svc.cluster.local:9000"),
									},
								},
								ImagePullPolicy: pulumi.String("Always"),
							},
						},
					},
				},
			},
		}, dependencies...)
		if err != nil {
			return err
		}

		err = createDaprMonitoring(ctx)
		if err != nil {
			return err
		}

		for _, service := range services {
			_, err := NewServiceDeployment(ctx, service, dependencies...)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func createOpenFunction(ctx *pulumi.Context, dapr *helm.Chart) (*helm.Release, error) {
	openfunctionNamespace, err := corev1.NewNamespace(ctx, "openfunction", &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{Name: pulumi.String("openfunction")},
	})
	if err != nil {
		return nil, err
	}

	openfunction, err := helm.NewRelease(ctx, "openfunction", &helm.ReleaseArgs{
		Name:      pulumi.String("openfunction"),
		Version:   pulumi.String("0.5.0"),
		Chart:     pulumi.String("openfunction"),
		Namespace: openfunctionNamespace.Metadata.Elem().Name().Elem(),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://openfunction.github.io/charts/"),
		},
		Values: pulumi.Map{
			"global": pulumi.Map{
				"Contour": pulumi.Map{
					"enabled": pulumi.Bool(false),
				},
				"KnativeServing": pulumi.Map{
					"enabled": pulumi.Bool(false),
				},
				"Dapr": pulumi.Map{
					"enabled": pulumi.Bool(false),
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{openfunctionNamespace}), pulumi.DependsOnInputs(dapr.Ready))
	if err != nil {
		return nil, err
	}
	return openfunction, nil
}

func createRedis(ctx *pulumi.Context) (*helm.Chart, error) {
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
		return nil, err
	}
	return redis, nil
}

func createMinio(ctx *pulumi.Context) (*helm.Chart, error) {
	minioNamespace, err := corev1.NewNamespace(ctx, "minio", &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{Name: pulumi.String("minio")},
	})
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return minio, nil
}

func createDapr(ctx *pulumi.Context) (*helm.Chart, error) {
	daprNamespace, err := corev1.NewNamespace(ctx, "dapr-system", &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{Name: pulumi.String("dapr-system")},
	})
	if err != nil {
		return nil, err
	}

	dapr, err := helm.NewChart(ctx, "dapr", helm.ChartArgs{
		Version:   pulumi.String("1.10"),
		Chart:     pulumi.String("dapr"),
		Namespace: daprNamespace.Metadata.Elem().Name().Elem(),
		FetchArgs: &helm.FetchArgs{
			Repo: pulumi.String("https://dapr.github.io/helm-charts/"),
		},
		Values: pulumi.Map{
			"global": pulumi.Map{
				"logAsJson": pulumi.Bool(true),
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{daprNamespace}), pulumi.Transformations([]pulumi.ResourceTransformation{
		// Source: https://www.pulumi.com/registry/packages/kubernetes/how-to-guides/managing-resources-with-server-side-apply/#helm-charts
		// Ignore changes that will be overwritten by the kruise-manager deployment.
		func(args *pulumi.ResourceTransformationArgs) *pulumi.ResourceTransformationResult {
			if args.Type == "kubernetes:admissionregistration.k8s.io/v1:ValidatingWebhookConfiguration" ||
				args.Type == "kubernetes:admissionregistration.k8s.io/v1:MutatingWebhookConfiguration" {
				return &pulumi.ResourceTransformationResult{
					Props: args.Props,
					Opts: append(args.Opts, pulumi.IgnoreChanges([]string{
						"metadata.annotations.template",
						"webhooks[*].clientConfig",
					})),
				}
			}

			if args.Name == "dapr-system/dapr-webhook-ca" ||
				args.Name == "dapr-system/dapr-webhook-cert" ||
				args.Name == "dapr-system/dapr-sidecar-injector-cert" {
				return &pulumi.ResourceTransformationResult{
					Props: args.Props,
					Opts: append(args.Opts, pulumi.IgnoreChanges([]string{
						"data",
					})),
				}
			}
			return nil
		},
	}))
	if err != nil {
		return nil, err
	}
	return dapr, nil
}

func createDaprMonitoring(ctx *pulumi.Context) error {
	daprMonitoring, err := corev1.NewNamespace(ctx, "dapr-monitoring", &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{Name: pulumi.String("dapr-monitoring")},
	})
	if err != nil {
		return err
	}

	_, err = helm.NewChart(ctx, "prometheus", helm.ChartArgs{
		Version:   pulumi.String("20.2.1"),
		Chart:     pulumi.String("prometheus"),
		Namespace: daprMonitoring.Metadata.Elem().Name().Elem(),
		FetchArgs: &helm.FetchArgs{
			Repo: pulumi.String("https://prometheus-community.github.io/helm-charts"),
		},
		Values: pulumi.Map{
			"server": pulumi.Map{
				"persistentVolume": pulumi.Map{
					"enabled": pulumi.Bool(false),
				},
			},
			"alertmanager": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
			"prometheus-pushgateway": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
			"prometheus-node-exporter": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
			"kube-state-metrics": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
			"configmapReload": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{daprMonitoring}))
	if err != nil {
		return err
	}

	_, err = helm.NewChart(ctx, "grafana", helm.ChartArgs{
		Version:   pulumi.String("6.54.0"),
		Chart:     pulumi.String("grafana"),
		Namespace: daprMonitoring.Metadata.Elem().Name().Elem(),
		FetchArgs: &helm.FetchArgs{
			Repo: pulumi.String("https://grafana.github.io/helm-charts"),
		},
		Values: pulumi.Map{
			"adminPassword": pulumi.String("admin"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{daprMonitoring}))
	if err != nil {
		return err
	}

	_, err = helm.NewChart(ctx, "elasticsearch", helm.ChartArgs{
		Version:   pulumi.String("7.17.3"),
		Chart:     pulumi.String("elasticsearch"),
		Namespace: daprMonitoring.Metadata.Elem().Name().Elem(),
		FetchArgs: &helm.FetchArgs{
			Repo: pulumi.String("https://helm.elastic.co"),
		},
		Values: pulumi.Map{
			"replicas": pulumi.Int(1),
			"persistence": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{daprMonitoring}))
	if err != nil {
		return err
	}

	_, err = helm.NewChart(ctx, "kibana", helm.ChartArgs{
		Version:   pulumi.String("7.17.3"),
		Chart:     pulumi.String("kibana"),
		Namespace: daprMonitoring.Metadata.Elem().Name().Elem(),
		FetchArgs: &helm.FetchArgs{
			Repo: pulumi.String("https://helm.elastic.co"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{daprMonitoring}))
	if err != nil {
		return err
	}

	_, err = yaml.NewConfigGroup(ctx, "logging-components", &yaml.ConfigGroupArgs{
		Files: []string{filepath.Join("logging-components", "*.yaml")},
	})
	if err != nil {
		return err
	}

	return nil
}
