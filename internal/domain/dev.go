package domain

import (
	"github.com/ATenderholt/lambda-router/settings"
	aws "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"path/filepath"
)

// DevFunction contains the settings to support development without deploying directly
// to lambda-router with AWS CLI / Terraform
type DevFunction struct {
	name        string
	Handler     string
	Runtime     string
	BasePath    string `yaml:"basePath"`
	Environment []string
	DepPath     string
}

func (d DevFunction) Name() string {
	return d.name
}

func (d *DevFunction) SetName(name string) {
	d.name = name
}

func (d DevFunction) EnvVars() []string {
	environment := make([]string, 2+len(d.Environment))
	environment[0] = "DOCKER_LAMBDA_STAY_OPEN=1"
	environment[1] = "DOCKER_LAMBDA_WATCH=1"

	for i, env := range d.Environment {
		environment[i+2] = env
	}

	return environment
}

func (d DevFunction) HandlerCmd() []string {
	return []string{d.Handler}
}

func (d DevFunction) AwsRuntime() aws.Runtime {
	return aws.Runtime(d.Runtime)
}

func (d DevFunction) GetDestPath(cfg *settings.Config) string {
	if filepath.IsAbs(d.BasePath) {
		return d.BasePath
	}

	base, err := filepath.Abs(cfg.DevConfigFile)
	if err != nil {
		panic(err)
	}

	return filepath.Join(filepath.Dir(base), d.BasePath)
}

func (d DevFunction) GetLayerDestPath(cfg *settings.Config) string {
	return d.DepPath
}
