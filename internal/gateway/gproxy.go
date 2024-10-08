package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/dlshle/dockman/pkg/dockerx"
	"github.com/dlshle/dockman/pkg/gproxy"
	"github.com/dlshle/gommon/slices"
	"gopkg.in/yaml.v2"
)

const (
	gproxyLabel   = "v-gproxy-gateway"
	gproxyAPIPort = 17768
)

// GProxy supprts layer-4 proxies
type GProxyGateway struct {
}

func NewGProxyGateway() GatewayStrategy {
	return &GProxyGateway{}
}

func (g *GProxyGateway) BackendContainersByNetwork(ctx context.Context, dc *dockerx.DockerClient, network string) ([]*dockerx.Container, error) {
	return dc.ListContainers(ctx, map[string]string{"network": network, "label": fmt.Sprintf("gateway=%s", gproxyLabel)})
}

// CurrentConfig implements GatewayStrategy.
func (g *GProxyGateway) CurrentConfig(ctx context.Context, dc *dockerx.DockerClient, appName string, networkName string) (*GatewayDeploymentConfig, error) {
	container, err := g.GatewayContainerByAppName(ctx, dc, appName)
	if err != nil {
		return nil, err
	}
	gatewayIP, exists := container.IPAddresses[networkName]
	if !exists {
		return nil, fmt.Errorf("container %s does not have IP address in network %s", container.ID, networkName)
	}
	cfgResp, err := http.Get(fmt.Sprintf("http://%s:%d/config", gatewayIP, gproxyAPIPort))
	if err != nil {
		return nil, err
	}
	if cfgResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %s", cfgResp.Status)
	}
	cfgData, err := io.ReadAll(cfgResp.Body)
	if err != nil {
		return nil, err
	}
	cfg, err := gproxy.UnmarshalConfig(cfgData)
	if err != nil {
		return nil, err
	}

	backendsMap := make(map[string]bool)
	portsMap := make(map[int]bool)
	for _, l := range cfg.Upstreams {
		for _, b := range l.Backends {
			backendsMap[b.Host] = true
			portsMap[b.Port] = true
		}
	}

	depCfg := &GatewayDeploymentConfig{
		BackendContainerNames: make([]string, 0),
		Ports:                 make([]string, 0),
		Network:               appName,
		ExposedPorts:          make([]*ExposedPort, 0),
	}

	for b := range backendsMap {
		depCfg.BackendContainerNames = append(depCfg.BackendContainerNames, b)
	}
	for p := range portsMap {
		depCfg.Ports = append(depCfg.Ports, strconv.Itoa(p))
	}

	for private, public := range container.ExposedPorts {
		depCfg.ExposedPorts = append(depCfg.ExposedPorts, &ExposedPort{
			Exposed: strconv.Itoa(int(public)),
			Port:    strconv.Itoa(int(private)),
		})
	}

	return UnmarshalGatewayDeploymentConfig(cfgData)
}

// DeployGatewayContainer implements GatewayStrategy.
func (g *GProxyGateway) DeployGatewayContainer(ctx context.Context, dc *dockerx.DockerClient, cfg *GatewayDeploymentConfig) error {
	portMapping := make(map[string]string)
	for _, ep := range cfg.ExposedPorts {
		portMapping[ep.Port] = ep.Exposed
	}
	gproxyCfgData, err := g.marshalledGProxyConfig(cfg)
	if err != nil {
		return err
	}
	env := fmt.Sprintf(`GPROXY_CONFIG=%s`, string(gproxyCfgData))
	runOpts := &dockerx.RunOptions{
		Image:         "gproxy",
		ContainerName: cfg.AppName,
		Detached:      true,
		Networks:      []string{cfg.Network},
		PortMapping:   portMapping,
		Envs:          []string{env},
	}
	_, err = dc.RunImage(ctx, runOpts)
	return err
}

func (g *GProxyGateway) marshalledGProxyConfig(cfg *GatewayDeploymentConfig) ([]byte, error) {
	gproxyCfg, err := g.gproxyConfig(cfg)
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(gproxyCfg)
}

func (g *GProxyGateway) gproxyConfig(cfg *GatewayDeploymentConfig) (*gproxy.Config, error) {
	gcfg := &gproxy.Config{
		Upstreams: make([]*gproxy.ListenerCfg, 0),
	}
	for _, port := range cfg.Ports {
		portVal, err := strconv.Atoi(port)
		if err != nil {
			return nil, err
		}
		upstream := &gproxy.ListenerCfg{
			Port:     portVal,
			Policy:   "roundrobin",
			Protocol: "tcp",
			Backends: slices.Map(cfg.BackendContainerNames, func(s string) *gproxy.Backend {
				return &gproxy.Backend{
					Host: s,
					Port: portVal,
				}
			}),
		}
		gcfg.Upstreams = append(gcfg.Upstreams, upstream)
	}
	return gcfg, nil
}

// GatewayContainerByAppName implements GatewayStrategy.
func (g *GProxyGateway) GatewayContainerByAppName(ctx context.Context, dc *dockerx.DockerClient, appName string) (*dockerx.Container, error) {
	containers, err := dc.ListContainers(ctx, map[string]string{"name": appName})
	if err != nil {
		return nil, err
	}
	if len(containers) == 0 {
		return nil, ErrGatewayNotDeployed
	}
	if len(containers) != 1 {
		return nil, fmt.Errorf("expected exactly one nginx container, got %d", len(containers))
	}
	return containers[0], nil
}

// Labels implements GatewayStrategy.
func (g *GProxyGateway) Labels() map[string]string {
	return map[string]string{"gateway": gproxyLabel}
}

// ReloadGatewayContainer implements GatewayStrategy.
func (g *GProxyGateway) ReloadGatewayContainer(ctx context.Context, dc *dockerx.DockerClient, cfg *GatewayDeploymentConfig) error {
	container, err := g.GatewayContainerByAppName(ctx, dc, cfg.AppName)
	if err != nil {
		return err
	}
	ipAddr := container.IPAddresses[cfg.Network]

	gproxyCfgData, err := g.marshalledGProxyConfig(cfg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("http://%s:%d/config", ipAddr, gproxyAPIPort), bytes.NewReader(gproxyCfgData))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %s", resp.Status)
	}
	return nil
}
