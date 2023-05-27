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

type Test mg.Namespace

// Unit Test all unit tests
func (t Test) Unit() error {
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

// IntegrationStatus Integration test for the Status Service
func (t Test) IntegrationStatus() error {
	//TODO: Doesnt work anymore :( dagger hangs
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

// ComponentStatus Component test for the status service
func (t Test) ComponentStatus() error {
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
		WithServiceBinding("redis", redis).
		WithFile("status/config/config.yaml", directory.File("./status/config/config.yaml")).
		WithFile("status_service", output).
		// Run actual service with dapr sidecar
		WithExec(runStatusWithDapr).
		WithExposedPort(3500)

	integrationTest := baseContainer.
		WithServiceBinding("status-dapr", status).
		WithEnvVariable("STATUS_SERVICE_ADDR", "http://status-dapr:3500").
		WithExec([]string{"go", "test", "-v", "status/cmd/component_test.go"})

	stdout, err := integrationTest.Stdout(context.Background())
	fmt.Println(stdout)
	return err
}

func (t Test) E2E() error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	directory := client.Host().Directory(".")

	redis := client.Container().
		From("redis").
		WithExposedPort(6379).
		WithExec(nil)

	krakendConfig := krakendContainer(client, "local").File("krakend.json")
	krakendJWK := krakendContainer(client, "local").File("/usr/jwk_private_key.json")

	e2eTest := client.Container().From("ubuntu:18.04").
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "-y", "wget", "curl"}).
		// install krakend and go
		WithExec([]string{"wget", "-q", "https://repo.krakend.io/bin/krakend_2.3.2_amd64_generic-linux.tar.gz"}).
		WithExec([]string{"tar", "-C", "/usr/local", "-xzf", "krakend_2.3.2_amd64_generic-linux.tar.gz"}).
		WithExec([]string{"wget", "-q", "https://go.dev/dl/go1.20.4.linux-amd64.tar.gz"}).
		WithExec([]string{"tar", "-C", "/usr/local", "-xzf", "go1.20.4.linux-amd64.tar.gz"}).
		WithEnvVariable("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/usr/local/usr/bin").
		WithFile("/tmp/krakend.json", krakendConfig).
		WithFile("/usr/jwk_private_key.json", krakendJWK).
		// Prepare Dapr
		WithExec([]string{"wget", "-q", "https://raw.githubusercontent.com/dapr/cli/master/install/install.sh"}).
		WithExec([]string{"chmod", "+x", "install.sh"}).
		WithExec([]string{"./install.sh"}).
		WithExec([]string{"dapr", "init", "--slim"}).
		// Prepare Go Mod
		WithWorkdir("/usr/app/").
		WithFile("/usr/app/go.mod", directory.File("go.mod")).
		WithFile("/usr/app/go.sum", directory.File("go.sum")).
		WithExec([]string{"go", "mod", "download"}).
		WithMountedDirectory("", directory).
		WithServiceBinding("redis", redis).
		WithEnvVariable("NO_DAGGER_CACHE", time.Now().String()).
		// Run E2E
		WithExec([]string{"test-e2e/run.sh"})

	stdout, err := e2eTest.Stdout(context.Background())
	fmt.Println(stdout)
	return err
}

// E2eLocal all services must be running, for this test to be working
func (t Test) E2eLocal() error {
	args := []string{"run"}
	args = append(args, runDaprArgs("e2e", 9999, 9998)...)
	args = append(args, []string{"--", "go", "test", "test-e2e/e2e_test.go"}...)
	return sh.RunWithV(nil, "dapr", args...)
}
