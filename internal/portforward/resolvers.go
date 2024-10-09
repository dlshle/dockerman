package portforward

import (
	"context"
	"errors"
	"strings"

	"github.com/dlshle/dockman/pkg/dockerx"
	"github.com/dlshle/dockman/pkg/proxy/server"
	"github.com/dlshle/gommon/logging"
)

func CreatePortforwardBackendResolver(docker *dockerx.DockerClient) server.HostResolver {
	containerIDResolver := func(containerID string) (string, error) {
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

	containerNameResolver := func(containerName string) (string, error) {
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
	return func(backendName string) (string, error) {
		resolveRequest := strings.Split(backendName, ":")
		if len(resolveRequest) == 1 {
			return server.PlainHostResolver(backendName)
		}
		switch resolveRequest[0] {
		case "containerId":
			return containerIDResolver(resolveRequest[1])
		case "containerName":
			return containerNameResolver(resolveRequest[1])
		default:
			return "", errors.New("invalid resolve request: " + backendName)
		}
	}
}
