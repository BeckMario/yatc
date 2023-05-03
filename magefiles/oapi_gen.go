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
)

type Generate mg.Namespace

type oapiGen struct {
	dir       string
	oapiPaths []paths
}

type paths struct {
	dir          string
	serverConfig config
	clientConfig *config
	spec         string
}

type config struct {
	path   string
	output string
}

func checkForVersions(apiDir string) bool {
	openapi := fmt.Sprintf("%s/openapi.yaml", apiDir)
	_, err := os.Stat(openapi)
	return os.IsNotExist(err)
}

func getVersions(apiDir string) ([]string, error) {
	dirEntries, err := os.ReadDir(apiDir)
	if err != nil {
		return nil, err
	}
	dirs := make([]string, 0)
	for _, entry := range dirEntries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	return dirs, nil
}

func newPaths(dir string, hasClient bool) (paths, error) {
	spec := fmt.Sprintf("%s/openapi.yaml", dir)
	serverConfigPath := fmt.Sprintf("%s/oapi-codegen-config.server.yaml", dir)
	clientConfigPath := fmt.Sprintf("%s/oapi-codegen-config.client.yaml", dir)

	serverConfig, err := newOapiConfig(serverConfigPath)
	if err != nil {
		return paths{}, err
	}

	var clientConfig *config
	if hasClient {
		c, err := newOapiConfig(clientConfigPath)
		clientConfig = &c
		if err != nil {
			return paths{}, err
		}
	}

	return paths{dir, serverConfig, clientConfig, spec}, nil
}

func newOapiGen(service string, hasClient bool) (oapiGen, error) {
	apiDir := fmt.Sprintf("%s/api-definition", service)

	oapiPaths := make([]paths, 0)

	if checkForVersions(apiDir) {
		versions, err := getVersions(apiDir)
		if err != nil {
			return oapiGen{}, err
		}

		for _, version := range versions {
			dir := fmt.Sprintf("%s/%s", apiDir, version)
			paths, err := newPaths(dir, hasClient)
			if err != nil {
				return oapiGen{}, err
			}
			oapiPaths = append(oapiPaths, paths)
		}
	} else {
		paths, err := newPaths(apiDir, hasClient)
		if err != nil {
			return oapiGen{}, err
		}
		oapiPaths = append(oapiPaths, paths)
	}

	return oapiGen{apiDir, oapiPaths}, nil

}

func oapiCodeGenBase(client *dagger.Client) *dagger.Container {
	return client.Container().
		From("golang:alpine3.17").
		WithExec([]string{"go", "install", "github.com/deepmap/oapi-codegen/cmd/oapi-codegen@master"}).
		WithMountedDirectory("/app", Repository(client)).
		WithWorkdir("/app")
}

func newOapiConfig(path string) (config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return config{}, err
	}

	var oapiConfig struct {
		Output string `yaml:"output"`
	}
	err = yaml.Unmarshal(bytes, &oapiConfig)
	if err != nil {
		return config{}, err
	}

	return config{
		path:   path,
		output: oapiConfig.Output,
	}, nil
}

func (paths *paths) hasBeenModified() (bool, error) {
	serverChange, err := target.Dir(paths.serverConfig.output, paths.dir)
	fmt.Printf("Source %s Destination %s modification? %t\n", paths.dir, paths.serverConfig.output, serverChange)
	if err != nil {
		return true, err
	}

	if paths.clientConfig != nil {
		clientChange, err := target.Dir((*paths.clientConfig).output, paths.dir)
		fmt.Printf("Source %s Destination %s modification? %t\n", paths.dir, (*paths.clientConfig).output, clientChange)

		if err != nil {
			return true, err
		}
		return serverChange || clientChange, nil
	}

	return serverChange, nil
}

func (oapiGen *oapiGen) generate(client *dagger.Client) error {
	for _, paths := range oapiGen.oapiPaths {
		modified, err := paths.hasBeenModified()
		if err != nil {
			return err
		}

		if modified {
			fmt.Println("Generate with oapi-codegen")
			if paths.clientConfig != nil {
				content, err := oapiCodeGenBase(client).
					WithExec([]string{"oapi-codegen", "-config", (*paths.clientConfig).path, paths.spec}).
					File((*paths.clientConfig).output).
					Contents(context.Background())
				if err != nil {
					return fmt.Errorf("api generate failed: %w", err)
				}

				err = os.WriteFile((*paths.clientConfig).output, []byte(content), 0o600)
				if err != nil {
					return fmt.Errorf("api generate failed: %w", err)
				}
			}

			content, err := oapiCodeGenBase(client).
				WithExec([]string{"oapi-codegen", "-config", paths.serverConfig.path, paths.spec}).
				File(paths.serverConfig.output).
				Contents(context.Background())
			if err != nil {
				return fmt.Errorf("api generate failed: %w", err)
			}

			err = os.WriteFile(paths.serverConfig.output, []byte(content), 0o600)
			if err != nil {
				return fmt.Errorf("api generate failed: %w", err)
			}
		}
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

	openapi, err := newOapiGen(service, needsClient)
	if err != nil {
		return err
	}

	err = openapi.generate(client)
	if err != nil {
		return err
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
