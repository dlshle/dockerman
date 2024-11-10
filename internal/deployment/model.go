package deployment

import (
	"github.com/dlshle/dockman/internal/config"
	"github.com/dlshle/dockman/internal/gateway"
)

type DeploymentBackend struct {
	ID    string
	Name  string
	Image string
}

type DeploymentGateway struct {
	Container *DeploymentBackend
	Config    *gateway.GatewayDeploymentConfig
}

type DeploymentInfo struct {
	AppConfig  *config.AppConfig
	Gateway    *DeploymentGateway
	Network    string
	Ports      []string
	Containers []*DeploymentBackend
}
