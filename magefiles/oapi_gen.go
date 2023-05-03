//go::build mage

package main

import (
	"context"
	"dagger.io/dagger"
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/target"
	"gopkg.in/yaml.v3"
	"os"
	"yatc/internal"
)

type Generate mg.Namespace

type oapiGenPaths struct {
	dir              string
	oapiServerConfig string
	serverOutput     string
	oapiClientConfig *string
	clientOutput     *string
	openapi          string
}
type oapiConfig struct {
	Output string `yaml:"output"`
}

func oapiCodeGenBase(client *dagger.Client) *dagger.Container {
	return client.Container().
		From("golang:alpine3.17").
		WithExec([]string{"go", "install", "github.com/deepmap/oapi-codegen/cmd/oapi-codegen@master"}).
		WithMountedDirectory("/app", Repository(client)).
		WithWorkdir("/app")
}

func newOapiConfig(path string) (*oapiConfig, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config oapiConfig
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func newOapiGenPaths(service string, hasClient bool) (*oapiGenPaths, error) {
	dir := fmt.Sprintf("%s/api-definition", service)
	openapi := fmt.Sprintf("%s/openapi.yaml", dir)

	oapiServerConfig := fmt.Sprintf("%s/oapi-codegen-config.server.yaml", dir)
	config, err := newOapiConfig(oapiServerConfig)
	if err != nil {
		return nil, err
	}
	serverOutput := config.Output

	var oapiClientConfig *string
	var clientOutput *string
	if hasClient {
		oapiClientConfig = internal.Ptr(fmt.Sprintf("%s/oapi-codegen-config.client.yaml", dir))
		config, err := newOapiConfig(*oapiClientConfig)
		if err != nil {
			return nil, err
		}
		clientOutput = &config.Output
	}

	return &oapiGenPaths{
		dir,
		oapiServerConfig,
		serverOutput,
		oapiClientConfig,
		clientOutput,
		openapi,
	}, nil
}

func (paths *oapiGenPaths) hasBeenModified() (bool, error) {
	serverChange, err := target.Dir(paths.serverOutput, paths.dir)
	fmt.Printf("Source %s Destination %s modification? %t\n", paths.dir, paths.serverOutput, serverChange)
	if err != nil {
		return true, err
	}

	if paths.clientOutput != nil {
		clientChange, err := target.Dir(*paths.clientOutput, paths.dir)
		fmt.Printf("Source %s Destination %s modification? %t\n", paths.dir, *paths.clientOutput, clientChange)

		if err != nil {
			return true, err
		}
		return serverChange || clientChange, nil
	}

	return serverChange, nil
}

func (paths *oapiGenPaths) generate(client *dagger.Client) error {
	if paths.oapiClientConfig != nil {
		content, err := oapiCodeGenBase(client).
			WithExec([]string{"oapi-codegen", "-config", *paths.oapiClientConfig, paths.openapi}).
			File(*paths.clientOutput).
			Contents(context.Background())
		if err != nil {
			return fmt.Errorf("api generate failed: %w", err)
		}

		err = os.WriteFile(*paths.clientOutput, []byte(content), 0o600)
		if err != nil {
			return fmt.Errorf("api generate failed: %w", err)
		}
	}

	content, err := oapiCodeGenBase(client).
		WithExec([]string{"oapi-codegen", "-config", paths.oapiServerConfig, paths.openapi}).
		File(paths.serverOutput).
		Contents(context.Background())
	if err != nil {
		return fmt.Errorf("api generate failed: %w", err)
	}

	err = os.WriteFile(paths.serverOutput, []byte(content), 0o600)
	if err != nil {
		return fmt.Errorf("api generate failed: %w", err)
	}

	return nil
}

// Service Generate chi server and client(if needed) for given services from openapi spec
func (Generate) Service(service string, needsClient bool) error {
	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	openapi, err := newOapiGenPaths(service, needsClient)
	if err != nil {
		return err
	}
	modified, err := openapi.hasBeenModified()
	if err != nil {
		return err
	}

	if modified {
		fmt.Println("Generate with oapi-codegen")
		if err := openapi.generate(client); err != nil {
			return err
		}
	}

	return nil
}

// All Generate all chi servers and clients(if needed) from openapi specs
func (Generate) All() error {
	services := []struct {
		name      string
		hasClient bool
	}{
		{"status", true},
		{"user", true},
		{"timeline", false},
		{"media", false},
	}

	fns := make([]interface{}, len(services))
	for i, service := range services {
		fn := mg.F(Generate.Service, service.name, service.hasClient)
		fns[i] = fn
	}
	mg.Deps(fns...)
	return nil
}
