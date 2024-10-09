package main

import (
	"context"
	"flag"
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

func main() {
	// dmproxy -h localhost -c d2123cf -p 5000:5000, 3000:3000, 80:80
	var (
		host             string
		containerName    string
		containerID      string
		plainHost        string
		portMappingsArgs string
	)

	flag.StringVar(&host, "h", "shop4u.ai:17688", "dockerman host address")
	flag.StringVar(&containerName, "cn", "", "docker container name")
	flag.StringVar(&containerID, "ci", "", "docker container id")
	flag.StringVar(&plainHost, "ph", "0.0.0.0", "proxy host address")
	flag.StringVar(&portMappingsArgs, "p", "5001:5000", "port mappings(source:dest), separated by comma")

	if err := validateInput(host, containerName, containerID, plainHost, portMappingsArgs); err != nil {
		log.Fatal(err)
	}

	portMappings, err := parsePortMappings(portMappingsArgs)
	if err != nil {
		log.Fatal(err)
	}

	backendName := plainHost
	if containerName != "" {
		backendName = "containerName:" + containerName
	} else if containerID != "" {
		backendName = "containerID:" + containerID
	}

	portForward := &PortForward{
		ProxyServerHost: host,
		BackendName:     backendName,
		PortMappings:    portMappings,
	}

	if err = listen(portForward); err != nil {
		log.Fatal(err)
	}
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
