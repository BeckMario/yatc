//go::build mage

package main

import (
	"context"
	"dagger.io/dagger"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"os"
)

type Docker mg.Namespace

// Up Start needed dependencies with docker compose
func (d Docker) Up() error {
	return sh.Run("docker-compose", "-f", "infrastructure/local/docker-compose.yaml", "up", "-d")
}

// Down Stop dependencies with docker compose
func (d Docker) Down() error {
	return sh.Run("docker-compose", "-f", "infrastructure/local/docker-compose.yaml", "down")
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
