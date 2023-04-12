package main

import (
	"encoding/json"
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
)

type Run mg.Namespace

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

type component struct {
	Metadata struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec struct {
		Type     string `yaml:"type"`
		Version  string `yaml:"version"`
		Metadata []struct {
			Name  string `yaml:"name"`
			Value string `yaml:"value"`
		} `yaml:"metadata"`
	} `yaml:"spec"`
}

type inputs struct {
	ComponentName string
	ComponentType string
	Uri           string
	Metadata      map[string]string
}

func (Run) MediaConversion() error {
	file, err := os.ReadFile("./components/pubsub.yaml")
	if err != nil {
		return err
	}
	var pubsubComponent component
	err = yaml.Unmarshal(file, &pubsubComponent)
	if err != nil {
		return err
	}

	pubsubMetadata := map[string]string{}

	for _, metadata := range pubsubComponent.Spec.Metadata {
		pubsubMetadata[metadata.Name] = metadata.Value
	}
	inputMap := map[string]inputs{"redis": {pubsubComponent.Metadata.Name, pubsubComponent.Spec.Type, "media", pubsubMetadata}}

	funcContext := struct {
		Name    string
		Version string
		Port    string
		Runtime string
		Inputs  map[string]inputs
	}{
		Name:    "HandleMessage",
		Version: "v1.0.0",
		Port:    "8084",
		Runtime: "Async",
		Inputs:  inputMap,
	}

	funcContextJson, err := json.Marshal(funcContext)
	if err != nil {
		return err
	}
	fmt.Println(string(funcContextJson))
	daprArgs := []string{"--app-id", "media-conversion", "--app-port", strconv.Itoa(8084),
		"--resources-path", "./components", "--app-protocol", "grpc", "-G", "50014"}
	dockerArgs := []string{"docker", "run", "--rm", "--env", fmt.Sprintf("FUNC_CONTEXT=%s", string(funcContextJson)),
		"--env", "CONTEXT_MODE=self-host", "--env", "DAPR_GRPC_PORT=50014", "--env", "APP_PROTOCOL=grpc", "--name", "media-conversion",
		"--network", "host", "media-conversion"}
	args := []string{"run"}
	args = append(args, daprArgs...)
	args = append(args, "--")
	args = append(args, dockerArgs...)
	return sh.RunWithV(nil, "dapr", args...)
}

// All Run all services
func (Run) All() {
	mg.Deps(Run.User, Run.Status, Run.Timeline, Run.Media)
}
