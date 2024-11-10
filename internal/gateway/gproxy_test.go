package gateway

import (
	"context"
	"slices"
	"testing"

	"github.com/dlshle/dockman/pkg/dockerx"
)

func TestGProxy(t *testing.T) {
	d, e := dockerx.NewDockerClient("tcp://192.168.0.158:2375")
	if e != nil {
		t.Errorf("failed to connect to docker daemon %v", e)
		t.FailNow()
	}
	g := NewGProxyGateway()
	ctx := context.Background()

	t.Run("test create gateway", func(t *testing.T) {
		containerName := "gproxy-gateway-test"
		network := "test-gproxy-network-abc"

		// data cleanup
		existingContainers, err := d.ListContainers(ctx, map[string]string{"name": containerName, "label": "gateway=" + gatewayLabel})
		if err != nil {
			t.Fatal(err)
		}
		if len(existingContainers) > 0 {
			for _, container := range existingContainers {
				if err = d.StopContainer(ctx, container.ID); err != nil {
					t.Fatal(err)
				}
				if err = d.RemoveContainer(ctx, container.ID); err != nil {
					t.Fatal(err)
				}
			}
		}
		gtwyContainer, err := g.GatewayContainerByAppName(ctx, d, network)
		if err != nil {
			if err != ErrGatewayNotDeployed {
				t.Fatal(err)
			}
		} else {
			if err = d.StopContainer(ctx, gtwyContainer.ID); err != nil {
				t.Fatal(err)
			}
			if err = d.RemoveContainer(ctx, gtwyContainer.ID); err != nil {
				t.Fatal(err)
			}
		}

		// actual test
		networks, err := d.ListNetworks(ctx, map[string]string{"name": network})
		if err != nil {
			t.Fatal(err)
		}
		if len(networks) == 0 {
			if _, err := d.CreateNetwork(ctx, network, nil); err != nil {
				t.Errorf("failed to create network %v", err)
				t.FailNow()
			}
		}

		containerID, err := d.RunImage(ctx, &dockerx.RunOptions{
			ContainerName: containerName,
			Image:         "config_center:v1",
			Detached:      true,
			Labels:        g.Labels(),
			Networks:      []string{network},
		})

		if err != nil {
			t.Errorf("failed to create container %s due to %v", containerName, err)
			t.FailNow()
		}

		cfg := &GatewayDeploymentConfig{
			AppName:               "sometestapp",
			BackendContainerNames: []string{containerName},
			Network:               network,
			Ports:                 []string{"80"},
			ExposedPorts:          []*ExposedPort{{"80", "80"}},
		}

		if err = g.DeployGatewayContainer(ctx, d, cfg); err != nil {
			t.Errorf("failed to deploy gateway container due to %v", err)
			t.FailNow()
		}
		// TODO another issue: container readiness probe is not working since we are requesting from a different machine
		// maybe use a readiness probe container within the same network? or use a master controller container to communicate and proxy readiness check requests?
		backendContainers, err := g.BackendContainersByNetwork(ctx, d, network)
		if err != nil {
			t.Errorf("failed to get backend container due to %v", err)
			t.FailNow()
		}
		if len(backendContainers) != 1 {
			t.Errorf("expected 1 backend container, got %d", len(backendContainers))
			t.FailNow()
		}
		if backendContainers[0].ID != containerID {
			t.Errorf("expected backend container id %s, got %s", containerID, backendContainers[0].ID)
			t.FailNow()
		}

		currCfg, err := g.CurrentConfig(ctx, d, cfg.AppName, cfg.Network)
		if err != nil {
			t.Errorf("failed to get current config due to %v", err)
			t.FailNow()
		}
		if !slices.Equal(currCfg.BackendContainerNames, []string{containerID}) {
			t.Errorf("expected backend container id %s, got %s", containerID, currCfg.BackendContainerNames)
			t.FailNow()
		}

		if err = d.StopContainer(ctx, containerID); err != nil {
			t.Errorf("failed to stop container %v due to %v", containerID, err)
			t.FailNow()
		}
		if err = d.RemoveContainer(ctx, containerID); err != nil {
			t.Errorf("failed to remove container %v due to %v", containerID, err)
			t.FailNow()
		}

	})
}
