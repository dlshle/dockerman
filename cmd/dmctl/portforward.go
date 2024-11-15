package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/dlshle/dockman/pkg/proxy/client"
)

type PortForward struct {
	ProxyServerHost string
	BackendName     string
	PortMappings    []*PortMapping
}

type PortMapping struct {
	Source int
	Dest   int
}

func listen(pf *PortForward) (err error) {
	var (
		wg sync.WaitGroup
	)
	for _, pm := range pf.PortMappings {
		wg.Add(1)
		go func(pm *PortMapping) {
			err = client.PortForward(context.Background(), &wg, pf.ProxyServerHost, pm.Source, &client.Remote{
				Host: pf.BackendName,
				Port: int32(pm.Dest),
			})
			if err != nil {
				log.Fatal(err)
			}
		}(pm)
	}
	wg.Wait()
	return
}

func validateInput(host, appName, containerID, plainHost, portMappingsArgs string) error {
	if host == "" {
		return fmt.Errorf("host is required")
	}
	if appName == "" && containerID == "" && plainHost == "" {
		return fmt.Errorf("container name or container ID or plain host is required")
	}
	if portMappingsArgs == "" {
		return fmt.Errorf("port mappings are required")
	}
	return nil
}

func parsePortMappings(portMappingsArgs string) ([]*PortMapping, error) {
	var portMappings []*PortMapping
	args := strings.Split(portMappingsArgs, ",")
	if len(args) == 0 {
		return nil, fmt.Errorf("invalid port mappings")
	}
	for _, arg := range args {
		ports := strings.Split(arg, ":")
		if len(ports) != 2 {
			return nil, fmt.Errorf("invalid port mapping: %s", arg)
		}
		source, err := strconv.Atoi(ports[0])
		if err != nil {
			return nil, fmt.Errorf("invalid source port: %s", arg[0])
		}
		dest, err := strconv.Atoi(ports[1])
		if err != nil {
			return nil, fmt.Errorf("invalid destination port: %s", arg[1])
		}
		portMappings = append(portMappings, &PortMapping{
			Source: source,
			Dest:   dest,
		})
	}
	return portMappings, nil
}
