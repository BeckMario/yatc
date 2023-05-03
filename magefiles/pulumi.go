//go::build mage

package main

import (
	"context"
	"dagger.io/dagger"
	"errors"
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Pulumi mg.Namespace

func getPulumiBase(client *dagger.Client) (*dagger.Container, error) {
	accessToken := os.Getenv("PULUMI_ACCESS_TOKEN")
	if accessToken == "" {
		return nil, errors.New("PULUMI_ACCESS_TOKEN environment variable needed")
	}

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		return nil, errors.New("KUBECONFIG environment variable needed")
	}

	infraDir := client.Host().Directory("./infrastructure")

	// Hacky way till c2h is implemented https://github.com/dagger/dagger/issues/4080
	//Connect daggerEngine to kind docker network, so pulumi can access the kubernetes cluster
	daggerEngineName, err := exec.Command("docker", "ps", "--filter", "name=^dagger-engine-*", "--format", "{{.Names}}").CombinedOutput()
	if err != nil {
		return nil, err
	}
	_ = sh.Run("docker", "network", "connect", "kind", strings.TrimSuffix(string(daggerEngineName), "\n"))

	fileBytes, err := os.ReadFile(kubeconfig)
	if err != nil {
		return nil, err
	}
	expression := regexp.MustCompile(`https://127.0.0.1:\d+`)
	newFileBytes := expression.ReplaceAll(fileBytes, []byte("https://yatc-control-plane:6443"))

	return client.Container().From("pulumi/pulumi-go").
		WithEnvVariable("PULUMI_ACCESS_TOKEN", accessToken).
		WithMountedDirectory("/app", infraDir).
		WithNewFile("/app/config", dagger.ContainerWithNewFileOpts{
			Contents: string(newFileBytes),
		}).
		WithEnvVariable("KUBECONFIG", "/app/config").
		WithWorkdir("/app/pulumi").
		WithExec([]string{"go", "mod", "download"}).
		WithExec([]string{"pulumi", "plugin", "install", "resource", "kubernetes", "v3.25.0"}), nil
}

func (p Pulumi) Preview() error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	base, err := getPulumiBase(client)
	if err != nil {
		return err
	}

	stdout, err := base.
		WithEnvVariable("NO_CACHE", time.Now().String()).
		WithExec([]string{"pulumi", "preview", "--stack", "dev", "--non-interactive"}).
		Stdout(context.Background())
	if err != nil {
		fmt.Println(stdout)
		return err
	}
	return nil
}

func (p Pulumi) Up() error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	base, err := getPulumiBase(client)
	if err != nil {
		return err
	}

	stdout, err := base.
		WithEnvVariable("NO_CACHE", time.Now().String()).
		WithExec([]string{"pulumi", "up", "--stack", "dev", "--non-interactive", "-y"}).
		Stdout(context.Background())
	if err != nil {
		fmt.Println(stdout)
		return err
	}
	return nil
}

func (p Pulumi) Destroy() error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	base, err := getPulumiBase(client)
	if err != nil {
		return err
	}

	stdout, err := base.
		WithEnvVariable("NO_CACHE", time.Now().String()).
		WithExec([]string{"pulumi", "destroy", "--stack", "dev", "--non-interactive", "-y"}).
		Stdout(context.Background())
	if err != nil {
		fmt.Println(stdout)
		return err
	}
	return nil
}
