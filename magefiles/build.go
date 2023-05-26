//go::build mage

package main

import (
	"context"
	"dagger.io/dagger"
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"os"
)

type Build mg.Namespace

func krakendContainer(client *dagger.Client, env string) *dagger.Container {
	krakendDir := client.Host().Directory("./infrastructure/krakend")

	return krakendDir.DockerBuild(dagger.DirectoryDockerBuildOpts{
		BuildArgs: []dagger.BuildArg{{Name: "ENV", Value: env}},
	})
}

func (b Build) krakend() error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	krakendImage := "reg.technicalonions.de/krakend-service:latest"
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
		krakendImage = "reg.technicalonions.de/krakend-service:local"
	}

	_, err = krakendContainer(client, env).Publish(context.Background(), krakendImage)
	if err != nil {
		return err
	}
	return nil
}

func (b Build) mediaConversion() error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	mediaConversionDir := client.Host().Directory("./media-conversion")

	//id := uuid.New()
	//runImageName := fmt.Sprintf("ttl.sh/%s:%s", id, "5m")
	runImageName := "reg.technicalonions.de/media-conversion-run-image:latest"
	_, err = mediaConversionDir.DockerBuild(dagger.DirectoryDockerBuildOpts{
		Dockerfile: "run.Dockerfile",
	}).Publish(context.Background(), runImageName)
	if err != nil {
		return err
	}

	//id = uuid.New()
	//imageName := fmt.Sprintf("ttl.sh/%s:%s", id, "5m")
	mediaConversionImageName := "reg.technicalonions.de/media-conversion:latest"
	buildCmd := []string{"build", mediaConversionImageName, "--publish", "--builder", "openfunction/builder-go:v2.4.0-1.17",
		"--run-image", runImageName, "--env", "FUNC_NAME=HandleMessage", "--env", "FUNC_CLEAR_SOURCE=true", "--path", "./media-conversion"}

	fmt.Println(buildCmd)

	return sh.Run("pack", buildCmd...)

	// Cant use because bug in dagger https://github.com/dagger/dagger/issues/4673
	/*	//Publish Temporary in ttl.sh anonymous registry, because i cant publish with pack, because there a no creds
		socket := client.Host().UnixSocket("/var/run/docker.sock")
		_, err = client.Container().From("buildpacksio/pack:latest").
			WithUser("root").
			WithUnixSocket("/var/run/docker.sock", socket).
			WithMountedDirectory("/workspace", mediaConversionDir).
			WithWorkdir("/workspace").
			WithExec(buildCmd).
			ExitCode(context.Background())
		if err != nil {
			return err
		}

		mediaConversionImageName := "reg.technicalonions.de/media-conversion:latest"
		_, err = client.Container().From(imageName).Publish(context.Background(), mediaConversionImageName)
		if err != nil {
			return err
		}*/
}

// Service Build and publish given service
func (b Build) Service(name string) error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	if name == "krakend" {
		return b.krakend()
	} else if name == "media-conversion" {
		return b.mediaConversion()
	}

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

// All Build and publish all services
func (b Build) All() error {
	services := []string{"status", "timeline", "media", "user", "login", "krakend", "media-conversion"}
	for _, service := range services {
		err := b.Service(service)
		if err != nil {
			return err
		}
	}
	return nil
}
