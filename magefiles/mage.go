//go::build mage

package main

import (
	"context"
	"dagger.io/dagger"
	"os"
)

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
