package portforward

import (
	"context"
	"errors"

	"github.com/dlshle/dockman/pkg/dockerx"
	"github.com/dlshle/dockman/pkg/proxy/server"
	"github.com/dlshle/gommon/logging"
)

func CreateContainerNameResolver(docker *dockerx.DockerClient) server.HostResolver {
	return func(containerName string) (string, error) {
		containers, err := docker.ListContainers(context.Background(), map[string]string{"name": containerName})
		if err != nil {
			return "", err
		}
		if len(containers) == 0 {
			return "", errors.New("container not found")
		}
		logging.GlobalLogger.Infof(context.Background(), "found %d containers by name %s, will use the first one", len(containers), containerName)
		for _, ip := range containers[0].IPAddresses {
			return ip, nil
		}
		return "", errors.New("no ip address found")
	}
}
