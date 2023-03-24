//go::build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"strconv"
)

type Run mg.Namespace
type Generate mg.Namespace

// All Generate all chi servers and clients(if needed) from openapi specs
func (Generate) All() error {
	services := []struct {
		name      string
		hasClient bool
	}{
		{"status", false},
		{"user", true},
		{"timeline", false},
	}

	fns := make([]interface{}, len(services))
	for i, service := range services {
		fn := mg.F(Generate.Service, service.name, service.hasClient)
		fns[i] = fn
	}
	mg.Deps(fns...)
	return nil
}

func runDapr(service string, appPort int, daprPort int) error {

	return sh.RunWithV(nil, "dapr", "run",
		"--app-id", service+"-service", "--app-port", strconv.Itoa(appPort), "--dapr-http-port", strconv.Itoa(daprPort),
		/*"--resources-path" , "../components" ,*/ "--", "go", "run", service+"/cmd/main.go")
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
