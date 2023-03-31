//go::build mage

package main

import (
	"context"
	"dagger.io/dagger"
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"os"
	"strconv"
	"time"
)

type Run mg.Namespace
type Generate mg.Namespace

func runDaprArgs(service string, appPort int, daprPort int) []string {
	return []string{"--app-id", service + "-service", "--app-port", strconv.Itoa(appPort),
		"--dapr-http-port", strconv.Itoa(daprPort), "--resources-path", "./components"}
}

func runDapr(service string, appPort int, daprPort int) error {
	args := []string{"run"}
	args = append(args, runDaprArgs(service, appPort, daprPort)...)
	args = append(args, []string{"--", "go", "run", service + "/cmd/main.go"}...)
	return sh.RunWithV(nil, "dapr", args...)
}

// Media Run service with dapr sidecar
func (Run) Media() error {
	mg.Deps(mg.F(Generate.Service, "media", false))
	return runDapr("media", 8083, 3503)
}

// Status Run service with dapr sidecar
func (Run) Status() error {
	mg.Deps(mg.F(Generate.Service, "status", false))
	return runDapr("status", 8082, 3500)
}

// User Run service with dapr sidecar
func (Run) User() error {
	mg.Deps(mg.F(Generate.Service, "user", true))
	return runDapr("user", 8080, 3502)

}

// Timeline Run service with dapr sidecar
func (Run) Timeline() error {
	mg.Deps(mg.F(Generate.Service, "timeline", false))
	return runDapr("timeline", 8081, 3501)
}

// All Run all services
func (Run) All() {
	mg.Deps(Run.User, Run.Status, Run.Timeline)
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

func Dagger() error {
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
