package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/dlshle/dockman/pkg/dockerx"
)

var (
	ErrGatewayNotDeployed = fmt.Errorf("gateway not deployed")
)

var (
	StrategyRegistry = map[string]GatewayStrategy{
		"nginx":  NewNginxGateway(),
		"gproxy": NewGProxyGateway(),
	}
)

type GatewayStrategy interface {
	CurrentConfig(ctx context.Context, dc *dockerx.DockerClient, appName string, networkName string) (*GatewayDeploymentConfig, error)
	GatewayContainerByAppName(ctx context.Context, dc *dockerx.DockerClient, appName string) (*dockerx.Container, error)
	BackendContainersByNetwork(ctx context.Context, dc *dockerx.DockerClient, network string) ([]*dockerx.Container, error)
	ReloadGatewayContainer(ctx context.Context, dc *dockerx.DockerClient, cfg *GatewayDeploymentConfig) error
	DeployGatewayContainer(ctx context.Context, dc *dockerx.DockerClient, cfg *GatewayDeploymentConfig) error
	Labels() map[string]string
}

type ExposedPort struct {
	Port    string
	Exposed string
}

type GatewayDeploymentConfig struct {
	AppName               string
	BackendContainerNames []string
	Network               string
	Ports                 []string // TODO change to AppConfig.PortConfig
	ExposedPorts          []*ExposedPort
}

func (cfg *GatewayDeploymentConfig) Json() string {
	jsonData, _ := json.Marshal(cfg)
	return string(jsonData)
}

func (cfg *GatewayDeploymentConfig) HasBackend(containerName string) bool {
	return slices.Contains(cfg.BackendContainerNames, containerName)
}

func (cfg *GatewayDeploymentConfig) RemoveBackend(containerName string) {
	cfg.BackendContainerNames = slices.DeleteFunc(cfg.BackendContainerNames, func(s string) bool {
		return s == containerName
	})
}

func UnmarshalGatewayDeploymentConfig(data []byte) (*GatewayDeploymentConfig, error) {
	var cfg GatewayDeploymentConfig
	err := json.Unmarshal(data, &cfg)
	return &cfg, err
}
