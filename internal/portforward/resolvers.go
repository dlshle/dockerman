package portforward

import (
	"context"
	"errors"

	"github.com/dlshle/dockman/pkg/dockerx"
	"github.com/dlshle/dockman/pkg/proxy/server"
)

func CreateContainerIDResolver(docker *dockerx.DockerClient) server.HostResolver {
	return func(containerID string) (string, error) {
		containers, err := docker.ListContainers(context.Background(), map[string]string{"id": containerID})
		if err != nil {
			return "", err
		}
		if len(containers) == 0 {
			return "", errors.New("container not found")
		}
		for _, ip := range containers[0].IPAddresses {
			return ip, nil
		}
		return "", errors.New("no ip address found")
	}
}
