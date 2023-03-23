package main

import (
	"fmt"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"gopkg.in/yaml.v3"
	"os"
	"yatc/internal"
)

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

func (paths *oapiGenPaths) generate() error {
	if paths.oapiClientConfig != nil {
		err := sh.RunWith(nil, "oapi-codegen", "-config", *paths.oapiClientConfig, paths.openapi)
		if err != nil {
			return err
		}
	}
	return sh.RunWith(nil, "oapi-codegen", "-config", paths.oapiServerConfig, paths.openapi)
}

// Service Generate chi server and client(if needed) for given services from openapi spec
func (Generate) Service(service string, needsClient bool) error {

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
		if err := openapi.generate(); err != nil {
			return err
		}
	}

	return nil
}
