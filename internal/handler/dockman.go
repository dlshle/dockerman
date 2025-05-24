package handler

import (
	"context"
	"errors"

	"github.com/dlshle/dockman/internal/config"
	"github.com/dlshle/dockman/internal/deployment"
	"github.com/dlshle/dockman/internal/gateway"
	"github.com/dlshle/dockman/internal/portforward"
	"github.com/dlshle/dockman/pkg/dockerx"
	"github.com/dlshle/gommon/logging"
	"github.com/docker/docker/api/types/container"
)

type DockmanHandler struct {
	docker      *dockerx.DockerClient
	portforward *portforward.Portforward
	deployment  *deployment.Deployment
}

func NewDockmanHandler(docker *dockerx.DockerClient,
	deployment *deployment.Deployment,
	portforward *portforward.Portforward) *DockmanHandler {
	return &DockmanHandler{docker: docker, deployment: deployment, portforward: portforward}
}

func (s *DockmanHandler) Deploy(ctx context.Context, appCfg *config.AppConfig) error {
	if appCfg == nil {
		return errors.New("app config is empty")
	}
	if appCfg.GatewayStrategy == nil {
		appCfg.GatewayStrategy = new(string)
		*appCfg.GatewayStrategy = "nginx"
	}
	strategy := gateway.StrategyRegistry[*appCfg.GatewayStrategy]
	if strategy == nil {
		return errors.New("unknown gateway strategy: " + *appCfg.GatewayStrategy)
	}
	logging.GlobalLogger.Infof(ctx, "deploying app %s with gateway strategy %s", appCfg.Name, *appCfg.GatewayStrategy)
	return s.deployment.Deploy(ctx, strategy, appCfg)
}

func (s *DockmanHandler) Rollout(ctx context.Context, appCfg *config.AppConfig) error {
	if appCfg == nil {
		return errors.New("app config is empty")
	}
	if appCfg.GatewayStrategy == nil {
		appCfg.GatewayStrategy = new(string)
		*appCfg.GatewayStrategy = "nginx"
	}
	if appCfg.RestartPolicy == nil {
		appCfg.RestartPolicy = new(string)
		*appCfg.RestartPolicy = string(container.RestartPolicyUnlessStopped)
	}
	strategy := gateway.StrategyRegistry[*appCfg.GatewayStrategy]
	if strategy == nil {
		return errors.New("unknown gateway strategy: " + *appCfg.GatewayStrategy)
	}
	logging.GlobalLogger.Infof(ctx, "deploying app %s with gateway strategy %s", appCfg.Name, *appCfg.GatewayStrategy)
	return s.deployment.Rollout(ctx, strategy, appCfg)
}

func (s *DockmanHandler) ListDeployments(ctx context.Context) ([]string, error) {
	return s.deployment.List(ctx)
}

func (s *DockmanHandler) InfoDeployment(ctx context.Context, appName string) (*deployment.DeploymentInfo, error) {
	return s.deployment.Info(ctx, appName)
}

func (s *DockmanHandler) Delete(ctx context.Context, appName string) error {
	return s.deployment.Delete(ctx, appName)
}

func (s *DockmanHandler) GatewayStrategies() []string {
	var strategies []string
	for k := range gateway.StrategyRegistry {
		strategies = append(strategies, k)
	}
	return strategies
}

func (s *DockmanHandler) StartPortforwardServer() error {
	return s.portforward.Start()
}
