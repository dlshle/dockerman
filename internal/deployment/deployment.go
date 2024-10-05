package deployment

import (
	"context"
	"fmt"
	"slices"

	"github.com/dlshle/dockman/internal/config"
	"github.com/dlshle/dockman/internal/gateway"
	"github.com/dlshle/dockman/pkg/dockerx"
	"github.com/dlshle/gommon/http"
	"github.com/dlshle/gommon/logging"
	slicesx "github.com/dlshle/gommon/slices"
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

	gatewayContainer, err := strategy.GatewayContainerByNetwork(ctx, d.docker, network)
	if err != nil {
		// if no gateway container, just rollout
		if err == gateway.ErrGatewayNotDeployed {
			return d.Rollout(ctx, strategy, appCfg)
		}
		return err
	}
	err = d.docker.RemoveContainer(ctx, gatewayContainer.ID)
	if err != nil {
		return err
	}

	return d.Rollout(ctx, strategy, appCfg)
}

func (d *Deployment) rollingUpdate(ctx context.Context, strategy gateway.GatewayStrategy, appCfg *config.AppConfig) (err error) {
	network := networkNameByAppConfig(appCfg)
	ports := slicesx.Map(appCfg.Ports, func(p config.PortConfig) string { return fmt.Sprintf("%d", p.Source) })
	containers, err := strategy.BackendContainersByNetwork(ctx, d.docker, network)
	if err != nil {
		return err
	}

	containersMap := slicesx.ToMap(containers, func(c *dockerx.Container) (string, *dockerx.Container) {
		return c.Names[0], c
	})

	gatewayCfg, err := strategy.CurrentConfig(ctx, d.docker, network)
	if err != nil {
		if err != gateway.ErrGatewayNotDeployed {
			return err
		}
		// gateway not deployed, use empty gateway config with prefilled values for now
		gatewayCfg = &gateway.GatewayDeploymentConfig{
			BackendContainerNames: []string{},
			Network:               network,
			Ports:                 ports,
		}
	}

	// TODO: add saga to support rollback
	for replicaID := range appCfg.Replicas {
		containerName := containerNameByAppConfig(appCfg, replicaID)
		container := containersMap[containerName]
		if container.Image == appCfg.Image {
			continue
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
			logging.GlobalLogger.Errorf(ctx, "failed to deploy new container %s for %s due to %w", containerName, newContainerID, err)
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
				logging.GlobalLogger.Errorf(ctx, "failed to stop old container %v due to %w", container, err)
				return err
			}
			if err = d.docker.RenameContainer(ctx, container.ID, containerName+"-old"); err != nil {
				logging.GlobalLogger.Errorf(ctx, "failed to rename old container %v due to %w", container, err)

				// recover last container
				d.docker.StartContainer(ctx, container.ID)

				return err
			}
		}

		// rename new container(remove the -canary part)
		err = d.docker.RenameContainer(ctx, newContainerID, containerName)
		if err != nil {
			logging.GlobalLogger.Errorf(ctx, "failed to rename new container %v due to %w", newContainerID, err)

			// this recovering logic is bad, need refactoring
			// try to recover old container
			if rerr := d.docker.RenameContainer(ctx, container.ID, container.Names[0]); rerr == nil {
				d.docker.StartContainer(ctx, container.ID)
			}

			// try to purge new container
			if rerr := d.docker.StopContainer(ctx, newContainerID); rerr == nil {
				d.docker.RemoveContainer(ctx, newContainerID)
			}
			return fmt.Errorf("failed to rename new container %v due to %w", newContainerID, err)
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
	_, err = d.docker.CreateNetwork(ctx, networkName)
	return err
}

func networkNameByAppConfig(appCfg *config.AppConfig) string {
	return fmt.Sprintf("v-network-%s", appCfg.Name)
}

func containerNameByAppConfig(appCfg *config.AppConfig, replicaID int) string {
	return fmt.Sprintf("v-%s-%d", appCfg.Name, replicaID)
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
