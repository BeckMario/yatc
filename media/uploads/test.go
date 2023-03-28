package main

import (
	"context"
	"dagger.io/dagger"
	"fmt"
	"os"
	"strconv"
	"time"
)

func runDaprArgs(service string, appPort int, daprPort int) []string {
	return []string{"--app-id", service + "-service", "--app-port", strconv.Itoa(appPort), "--dapr-http-port", strconv.Itoa(daprPort)}
}

func main() {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		panic(err)
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
	status := client.Container().From("ubuntu:18.04").
		// Prepare Dapr
		WithExec([]string{"apt", "update"}).
		WithExec([]string{"apt", "install", "-y", "wget"}).
		WithExec([]string{"wget", "-q", "https://raw.githubusercontent.com/dapr/cli/master/install/install.sh"}).
		WithExec([]string{"chmod", "+x", "install.sh"}).
		WithExec([]string{"./install.sh"}).
		WithExec([]string{"dapr", "init", "--slim"}).
		// Prepare Status Client
		WithMountedDirectory("/test-components", directory.Directory("./status/cmd/test-components")).
		WithServiceBinding("db", postgres).
		WithServiceBinding("redis", redis).
		WithFile("status_service", output).
		// Run actual service with dapr sidecar
		WithExec(runStatusWithDapr).
		WithExposedPort(3500)

	integrationTest := baseContainer.
		WithServiceBinding("status-dapr", status).
		WithExec([]string{"go", "test", "-v", "status/cmd/integration_test.go"})

	fmt.Println(integrationTest.Stdout(context.Background()))
}
