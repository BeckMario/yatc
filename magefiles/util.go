package main

import "dagger.io/dagger"

func Repository(client *dagger.Client) *dagger.Directory {
	return client.Host().Directory(".")
}

func RepositoryGoCodeOnly(client *dagger.Client) *dagger.Directory {
	return client.Host().Directory(".", dagger.HostDirectoryOpts{
		Include: []string{
			"**/*.go",
			"**/go.mod",
			"**/go.sum",
		},
	})
}

func goBase(client *dagger.Client) *dagger.Container {
	repo := RepositoryGoCodeOnly(client)
	onlyGoMod := client.Directory().
		WithFile("go.mod", repo.File("go.mod")).
		WithFile("go.sum", repo.File("go.sum"))

	return client.Container().
		From("golang:alpine3.17").
		WithWorkdir("/app").
		WithMountedDirectory("/app", onlyGoMod).
		WithExec([]string{"go", "mod", "download"}).
		WithMountedDirectory("/app", repo)
}
