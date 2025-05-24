package deployment

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/dlshle/dockman/internal/config"
	"github.com/dlshle/dockman/internal/gateway"
	"github.com/dlshle/dockman/pkg/dockerx"
	"github.com/dlshle/gommon/http"
	"github.com/dlshle/gommon/logging"
	slicesx "github.com/dlshle/gommon/slices"
)

const (
	NetworkLabelGtwyStrategy = "dm-gateway-strategy"
)

type Deployment struct {
	docker         *dockerx.DockerClient
	backendClients map[string]http.HTTPClient
}

func NewDeployment(docker *dockerx.DockerClient) *Deployment {
	return &Deployment{
		docker:         docker,
		backendClients: make(map[string]http.HTTPClient),
	}
}

func (d *Deployment) Rollout(ctx context.Context, strategy gateway.GatewayStrategy, appCfg *config.AppConfig) error {
	err := d.createNetworkIfNotExists(ctx, appCfg)
	if err != nil {
		return err
	}
	return d.rollingUpdate(ctx, strategy, appCfg)
}

// 100% rollout
func (d *Deployment) Deploy(ctx context.Context, strategy gateway.GatewayStrategy, appCfg *config.AppConfig) error {
	// this is simple, just do old stuff cleanup and then rollout
	network := networkNameByAppConfig(appCfg)
	oldContainers, err := strategy.BackendContainersByNetwork(ctx, d.docker, network)
	if err != nil {
		return err
	}
	for _, container := range oldContainers {
		err := d.docker.RemoveContainer(ctx, container.ID)
		if err != nil {
			return err
		}
	}

	gatewayContainer, err := strategy.GatewayContainerByAppName(ctx, d.docker, appCfg.Name)
	if err != nil {
		// if no gateway container, just rollout
		if err == gateway.ErrGatewayNotDeployed {
			return d.Rollout(ctx, strategy, appCfg)
		}
		return err
	}
	if err = d.docker.StopContainer(ctx, gatewayContainer.ID); err != nil {
		return err
	}
	if err = d.docker.RemoveContainer(ctx, gatewayContainer.ID); err != nil {
		return err
	}

	return d.Rollout(ctx, strategy, appCfg)
}

func (d *Deployment) List(ctx context.Context) ([]string, error) {
	var appNames []string
	networks, err := d.docker.ListNetworks(ctx, nil)
	if err != nil {
		logging.GlobalLogger.Errorf(ctx, "failed to list networks: %v", err)
		return nil, err
	}

	for _, network := range networks {
		if appName, hasPrefix := strings.CutPrefix(network.Name, "dm-network"); hasPrefix {
			appNames = append(appNames, appName)
		}
	}

	return appNames, nil
}

func (d *Deployment) Info(ctx context.Context, appName string) (*DeploymentInfo, error) {
	networkName := networkNameByAppName(appName)

	// network is only for validation
	networks, err := d.docker.ListNetworks(ctx, map[string]string{"name": networkName})
	if err != nil {
		return nil, err
	}

	if len(networks) == 0 {
		return nil, errors.New("app deployment for " + appName + " is not found")
	}

	gtwyStrategy := ""
	for _, network := range networks {
		if stg, exists := network.Labels[NetworkLabelGtwyStrategy]; exists {
			gtwyStrategy = stg
		}
	}

	strategy := gateway.StrategyRegistry[gtwyStrategy]
	if strategy == nil {
		return nil, fmt.Errorf("can not find gateway strategy from app name: %s", appName)
	}

	gtwyCfg, err := strategy.CurrentConfig(ctx, d.docker, appName, networkName)
	if err != nil {
		return nil, err
	}

	gatewayContainer, err := strategy.GatewayContainerByAppName(ctx, d.docker, appName)
	if err != nil {
		return nil, err
	}

	// backend containers
	backendContainers, err := strategy.BackendContainersByNetwork(ctx, d.docker, networkName)
	if err != nil {
		return nil, err
	}

	appCfg := &config.AppConfig{
		GatewayStrategy: &gtwyStrategy,
		Image:           backendContainers[0].Image,
		Name:            appName,
		Replicas:        len(backendContainers),
		RestartPolicy:   &backendContainers[0].RestartPolicy,
	}

	return &DeploymentInfo{
		AppConfig: appCfg,
		Gateway: &DeploymentGateway{
			Config: gtwyCfg,
			Container: &DeploymentBackend{
				ID:    gatewayContainer.ID,
				Image: gatewayContainer.Image,
				Name:  gatewayContainer.Names[0],
			},
		},
		Network: networkName,
		Containers: slicesx.Map(backendContainers, func(c *dockerx.Container) *DeploymentBackend {
			return &DeploymentBackend{
				ID:    c.ID,
				Image: c.Image,
				Name:  c.Names[0],
			}
		}),
	}, nil
}

func (d *Deployment) Delete(ctx context.Context, appName string) error {
	// network
	networkName := networkNameByAppName(appName)

	networks, err := d.docker.ListNetworks(ctx, map[string]string{"name": networkName})
	if err != nil {
		return err
	}

	if len(networks) == 0 {
		return errors.New("app deployment for " + appName + " is not found")
	}

	// containers
	containers, err := d.docker.ListContainers(ctx, map[string]string{"network": networkName})
	if err != nil {
		return err
	}
	for _, container := range containers {
		if container.State == "running" {
			if err = d.docker.StopContainer(ctx, container.ID); err != nil {
				return err
			}
		}
		if err := d.docker.RemoveContainer(ctx, container.ID); err != nil {
			logging.GlobalLogger.Errorf(ctx, "failed to remove container %s: %v", container.ID, err)
			return err
		}
	}

	if len(networks) > 0 {
		// delete networks
		for _, n := range networks {
			if err = d.docker.RemoveNetwork(ctx, n.Name); err != nil {
				logging.GlobalLogger.Errorf(ctx, "failed to remove network %s: %v", n.Name, err)
				return err
			}
		}
	}
	return nil
}

func (d *Deployment) rollingUpdate(ctx context.Context, strategy gateway.GatewayStrategy, appCfg *config.AppConfig) (err error) {
	network := networkNameByAppConfig(appCfg)
	ports := slicesx.Map(appCfg.Ports, func(p config.PortConfig) string { return fmt.Sprintf("%d", p.Source) })

	// image existanec check
	images, err := d.docker.ListImages(ctx, map[string]string{"reference": appCfg.Image})
	if err != nil {
		return err
	}
	if len(images) == 0 {
		return fmt.Errorf("image %s not found", appCfg.Image)
	}
	image := images[0]

	containers, err := strategy.BackendContainersByNetwork(ctx, d.docker, network)
	if err != nil {
		return err
	}

	containersMap := slicesx.ToMap(containers, func(c *dockerx.Container) (string, *dockerx.Container) {
		logging.GlobalLogger.Infof(ctx, "existing backend container(%s): %v", c.Names[0], c)
		return c.Names[0], c
	})

	gatewayCfg, err := strategy.CurrentConfig(ctx, d.docker, appCfg.Name, network)
	if err != nil {
		if err != gateway.ErrGatewayNotDeployed {
			return err
		}
		// gateway not deployed, use empty gateway config with prefilled values for now
		gatewayCfg = &gateway.GatewayDeploymentConfig{
			AppName:               appCfg.Name,
			BackendContainerNames: []string{},
			Network:               network,
			Ports:                 ports,
		}
		if err = strategy.DeployGatewayContainer(ctx, d.docker, gatewayCfg); err != nil {
			return fmt.Errorf("failed to initiate gateway: %v", err)
		}
	}

	// TODO: add saga to support rollback
	for replicaID := range appCfg.Replicas {
		containerName := containerNameByAppConfig(appCfg, replicaID)
		container := containersMap[containerName]
		if container != nil {
			logging.GlobalLogger.Infof(ctx, "container %s exists: %v", containerName, container)
		}
		if container != nil && container.ImageID == image.ID {
			continue
		}

		// remove existing canary containers
		canaryContainers, err := d.docker.ListContainers(ctx, map[string]string{"name": containerName + "-canary"})
		if err != nil {
			return err
		}
		for _, canaryContainer := range canaryContainers {
			if err = d.docker.ForceRemoveContainer(ctx, canaryContainer.ID); err != nil {
				return err
			}
		}

		// deploy new container with new image
		newContainerID, err := d.docker.RunImage(ctx, &dockerx.RunOptions{
			Image:         appCfg.Image,
			Networks:      []string{network},
			ContainerName: containerName + "-canary",
			Detached:      true,
			Labels:        strategy.Labels(),
		})
		if err != nil {
			logging.GlobalLogger.Errorf(ctx, "failed to deploy new container %s for %s due to %v", containerName, newContainerID, err)
			return err
		}
		logging.GlobalLogger.Infof(ctx, "new container %s(%s) for %s deployed", containerName, newContainerID, appCfg.Name)

		// rediness check
		if err = d.checkRediness(ctx, appCfg, replicaID); err != nil {
			logging.GlobalLogger.Infof(ctx, "rediness check failed for container %s(%s)", containerName, newContainerID)
			return err
		}

		// stop and rename old container
		if container != nil {
			if err = d.docker.StopContainer(ctx, container.ID); err != nil {
				logging.GlobalLogger.Errorf(ctx, "failed to stop old container %v due to %v", container, err)
				return err
			}
			if err = d.docker.RenameContainer(ctx, container.ID, containerName+"-old"); err != nil {
				logging.GlobalLogger.Errorf(ctx, "failed to rename old container %v due to %v", container, err)

				// recover last container
				d.docker.StartContainer(ctx, container.ID)

				return err
			}
		}

		// rename new container(remove the -canary part)
		err = d.docker.RenameContainer(ctx, newContainerID, containerName)
		if err != nil {
			logging.GlobalLogger.Errorf(ctx, "failed to rename new container %v due to %v", newContainerID, err)

			// this recovering logic is bad, need refactoring
			// try to recover old container
			if rerr := d.docker.RenameContainer(ctx, container.ID, container.Names[0]); rerr == nil {
				d.docker.StartContainer(ctx, container.ID)
			}

			// try to purge new container
			if rerr := d.docker.StopContainer(ctx, newContainerID); rerr == nil {
				d.docker.RemoveContainer(ctx, newContainerID)
			}
			return fmt.Errorf("failed to rename new container %v due to %v", newContainerID, err)
		}

		// setup gateway config with new container config(if not match and reload gateway)
		if !gatewayCfg.HasBackend(containerName) {
			gatewayCfg.BackendContainerNames = append(gatewayCfg.BackendContainerNames, containerName)
			if container != nil && gatewayCfg.HasBackend(container.Names[0]) {
				gatewayCfg.RemoveBackend(container.Names[0])
			}

			// reload gateway
			err = strategy.ReloadGatewayContainer(ctx, d.docker, gatewayCfg)
			if err != nil {
				return err
			}
		}

		// remove old container
		if container != nil {
			if err = d.docker.RemoveContainer(ctx, container.ID); err != nil {
				return err
			}
		}
	}
	// finally update gateway ports if not match
	if !slices.Equal(gatewayCfg.Ports, ports) {
		logging.GlobalLogger.Infof(ctx, "reloading gateway config for port update")
		gatewayCfg.Ports = ports
		return strategy.ReloadGatewayContainer(ctx, d.docker, gatewayCfg)
	}
	return nil
}

func (d *Deployment) createNetworkIfNotExists(ctx context.Context, appCfg *config.AppConfig) error {
	networkName := networkNameByAppConfig(appCfg)
	networks, err := d.docker.ListNetworks(ctx, nil)
	if err != nil {
		return err
	}
	for _, network := range networks {
		if network.Name == networkName {
			// network already exists
			return nil
		}
	}
	_, err = d.docker.CreateNetwork(ctx, networkName, map[string]string{NetworkLabelGtwyStrategy: *appCfg.GatewayStrategy})
	return err
}

func networkNameByAppConfig(appCfg *config.AppConfig) string {
	return networkNameByAppName(appCfg.Name)
}

func networkNameByAppName(appName string) string {
	return fmt.Sprintf("dm-network-%s", appName)
}

func containerNameByAppConfig(appCfg *config.AppConfig, replicaID int) string {
	return fmt.Sprintf("dm-%s-%d", appCfg.Name, replicaID)
}

func (d *Deployment) containerByReplicaID(ctx context.Context, appCfg *config.AppConfig, replicaID int) (*dockerx.Container, error) {
	containerName := containerNameByAppConfig(appCfg, replicaID)
	containers, err := d.docker.ListContainers(ctx, map[string]string{"name": containerName, "image": appCfg.Image})
	if err != nil {
		return nil, err
	}
	if len(containers) == 0 {
		return nil, fmt.Errorf("can not find container for app %s with replica id %d(%s)", appCfg.Name, replicaID, containerName)
	}
	if len(containers) > 1 {
		return nil, fmt.Errorf("found more than 1 containers for app %s with replica id %d(%s)", appCfg.Name, replicaID, containerName)
	}
	return containers[0], nil
}
