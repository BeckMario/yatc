package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
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

// All Run all services
func (Run) All() {
	mg.Deps(Run.User, Run.Status, Run.Timeline, Run.Media)
}
