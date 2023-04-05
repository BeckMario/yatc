//go::build mage

package main

import (
	"context"
	"dagger.io/dagger"
	"fmt"
	"github.com/magefile/mage/mg"
	"os"
	"time"
)

type Build mg.Namespace

func (b Build) Service(name string) error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

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
