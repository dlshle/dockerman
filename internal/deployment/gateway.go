package deployment

import (
	"context"
	"fmt"

	"github.com/dlshle/dockman/internal/config"
	"github.com/dlshle/dockman/internal/gateway"
	"github.com/dlshle/dockman/pkg/dockerx"
	"github.com/dlshle/gommon/slices"
)

func (d *Deployment) reloadGateway(ctx context.Context, strategy gateway.GatewayStrategy, network string, ports []config.PortConfig, backendContainers []*dockerx.Container) error {
	cfg := &gateway.GatewayDeploymentConfig{
		BackendContainerNames: slices.Map(backendContainers, func(c *dockerx.Container) string { return c.Names[0] }),
		Network:               network,
		Ports:                 slices.Map(ports, func(p config.PortConfig) string { return fmt.Sprintf("%d", p.Source) }),
	}
	return strategy.ReloadGatewayContainer(ctx, d.docker, cfg)
}
