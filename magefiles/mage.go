//go::build mage

package main

import (
	"context"
	"dagger.io/dagger"
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"os"
	"time"
)

type Build mg.Namespace

func (b Build) krakend() error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}

	krakendDir := client.Host().Directory("./infrastructure/krakend")

	krakendImage := "reg.technicalonions.de/krakend-service:latest"
	_, err = krakendDir.DockerBuild(dagger.DirectoryDockerBuildOpts{
		BuildArgs: []dagger.BuildArg{{Name: "ENV", Value: env}},
	}).Publish(context.Background(), krakendImage)
	if err != nil {
		return err
	}
	return nil
}

func (b Build) mediaConversion() error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	mediaConversionDir := client.Host().Directory("./media-conversion")

	//id := uuid.New()
	//runImageName := fmt.Sprintf("ttl.sh/%s:%s", id, "5m")
	runImageName := "reg.technicalonions.de/media-conversion-run-image:latest"
	_, err = mediaConversionDir.DockerBuild(dagger.DirectoryDockerBuildOpts{
		Dockerfile: "run.Dockerfile",
	}).Publish(context.Background(), runImageName)
	if err != nil {
		return err
	}

	//id = uuid.New()
	//imageName := fmt.Sprintf("ttl.sh/%s:%s", id, "5m")
	mediaConversionImageName := "reg.technicalonions.de/media-conversion:latest"
	buildCmd := []string{"build", mediaConversionImageName, "--publish", "--builder", "openfunction/builder-go:v2.4.0-1.17",
		"--run-image", runImageName, "--env", "FUNC_NAME=HandleMessage", "--env", "FUNC_CLEAR_SOURCE=true", "--path", "./media-conversion"}

	fmt.Println(buildCmd)

	return sh.Run("pack", buildCmd...)

	// Cant use because bug in dagger https://github.com/dagger/dagger/issues/4673
	/*	//Publish Temporary in ttl.sh anonymous registry, because i cant publish with pack, because there a no creds
		socket := client.Host().UnixSocket("/var/run/docker.sock")
		_, err = client.Container().From("buildpacksio/pack:latest").
			WithUser("root").
			WithUnixSocket("/var/run/docker.sock", socket).
			WithMountedDirectory("/workspace", mediaConversionDir).
			WithWorkdir("/workspace").
			WithExec(buildCmd).
			ExitCode(context.Background())
		if err != nil {
			return err
		}

		mediaConversionImageName := "reg.technicalonions.de/media-conversion:latest"
		_, err = client.Container().From(imageName).Publish(context.Background(), mediaConversionImageName)
		if err != nil {
			return err
		}*/
}

func (b Build) Service(name string) error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	if name == "krakend" {
		return b.krakend()
	} else if name == "media-conversion" {
		return b.mediaConversion()
	}

	file := goBase(client).WithExec([]string{"go", "build", "-o", fmt.Sprintf("build/%s_service", name), fmt.Sprintf("yatc/%s/cmd", name)}).
		File(fmt.Sprintf("build/%s_service", name))

	_, err = client.Container().
		WithFile("/app", file).
		WithEntrypoint([]string{"/app"}).
		Publish(context.Background(), fmt.Sprintf("reg.technicalonions.de/%s-service:latest", name))
	if err != nil {
		return err
	}

	return nil
}

func (b Build) All() error {
	services := []string{"status", "timeline", "media", "user"}
	for _, service := range services {
		err := b.Service(service)
		if err != nil {
			return err
		}
	}
	return nil
}

func Lint() error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}

	defer client.Close()
	_, err = client.Container().
		From("golangci/golangci-lint:v1.52-alpine").
		WithMountedDirectory("/app", RepositoryGoCodeOnly(client)).
		WithWorkdir("/app").
		WithExec([]string{"golangci-lint", "run", "-v", "--timeout", "5m"}).
		ExitCode(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func UnitTest() error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	stdout, err := goBase(client).
		WithExec([]string{"go", "test", "-v", "./..."}).
		Stdout(context.Background())
	if err != nil {
		err = fmt.Errorf("test failed: %w\n%s", err, stdout)
	}
	return err
}

func StatusIntegrationTest() error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	// Postgres
	postgres := client.Container().
		From("postgres").
		WithEnvVariable("POSTGRES_PASSWORD", "password").
		WithExposedPort(5432).
		WithExec(nil)

	stdout, err := goBase(client).
		WithServiceBinding("db", postgres).
		WithEnvVariable("DATABASE_CONNECTION_STRING", "postgres://postgres:password@db:5432/postgres?sslmode=disable").
		WithExec([]string{"go", "test", "-v", "status/internal/postgresrepo_test.go", "status/internal/repository.go"}).
		Stdout(context.Background())

	if err != nil {
		err = fmt.Errorf("test failed: %w\n%s", err, stdout)
	}
	return err
}

func StatusComponentTest() error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	directory := client.Host().Directory(".")
	baseContainer := client.Container().From("golang:latest").
		WithMountedDirectory("/src", directory).
		WithWorkdir("/src").
		WithExec([]string{"go", "mod", "download"}).
		WithEnvVariable("NO_DAGGER_CACHE", time.Now().String())

	buildContainer := baseContainer.WithExec([]string{"go", "build", "-o", "build/status_service", "yatc/status/cmd"})
	output := buildContainer.File("build/status_service")

	// Postgres
	postgres := client.Container().
		From("postgres").
		WithEnvVariable("POSTGRES_PASSWORD", "password").
		WithExposedPort(5432).
		WithExec(nil)

	// Redis. Status Service needs Redis as Message Broker
	redis := client.Container().
		From("redis").
		WithExposedPort(6379).
		WithExec(nil)

	runStatusWithDapr := []string{"dapr", "run"}
	runStatusWithDapr = append(runStatusWithDapr, runDaprArgs("status", 8082, 3500)...)
	runStatusWithDapr = append(runStatusWithDapr, "--resources-path", "./test-components")
	runStatusWithDapr = append(runStatusWithDapr, "--", "./status_service")

	// Prepare Status Client. Dapr is needed as Dependency
	status := client.Container().From("ubuntu:18.04").
		// Prepare Dapr
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "-y", "wget"}).
		WithExec([]string{"wget", "-q", "https://raw.githubusercontent.com/dapr/cli/master/install/install.sh"}).
		WithExec([]string{"chmod", "+x", "install.sh"}).
		WithExec([]string{"./install.sh"}).
		WithExec([]string{"dapr", "init", "--slim"}).
		// Prepare Service
		WithMountedDirectory("/test-components", directory.Directory("./status/config/test-components")).
		WithServiceBinding("db", postgres).
		WithServiceBinding("redis", redis).
		WithFile("status/config/config.yaml", directory.File("./status/config/config.yaml")).
		WithFile("status_service", output).
		WithEnvVariable("DATABASE", "postgres://postgres:password@db:5432/postgres?sslmode=disable").
		// Run actual service with dapr sidecar
		WithExec(runStatusWithDapr).
		WithExposedPort(3500)

	integrationTest := baseContainer.
		WithServiceBinding("status-dapr", status).
		WithEnvVariable("STATUS_SERVICE_ADDR", "http://status-dapr:3500").
		WithExec([]string{"go", "test", "-v", "status/cmd/integration_test.go"})

	stdout, err := integrationTest.Stdout(context.Background())
	fmt.Println(stdout)
	return err
}
