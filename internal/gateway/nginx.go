package gateway

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dlshle/dockman/pkg/dockerx"
)

const (
	gatewayLabel = "v-nginx-gateway"
	configPort   = 11451
)

type nginxGateway struct{}

func NewNginxGateway() GatewayStrategy {
	return &nginxGateway{}
}

func (n *nginxGateway) CurrentConfig(ctx context.Context, dc *dockerx.DockerClient, appName string, networkName string) (*GatewayDeploymentConfig, error) {
	container, err := n.GatewayContainerByAppName(ctx, dc, appName)
	if err != nil {
		return nil, err
	}
	gatewayIP, exists := container.IPAddresses[networkName]
	if !exists {
		return nil, fmt.Errorf("container %s does not have IP address in network %s", container.ID, appName)
	}
	cfgResp, err := http.Get(fmt.Sprintf("http://%s:%d/config", gatewayIP, configPort))
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
	return UnmarshalGatewayDeploymentConfig(cfgData)
}

func (n *nginxGateway) GatewayContainerByAppName(ctx context.Context, dc *dockerx.DockerClient, appName string) (*dockerx.Container, error) {
	containers, err := dc.ListContainers(ctx, map[string]string{"name": nginxContainerName(appName)})
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

func (n *nginxGateway) ReloadGatewayContainer(ctx context.Context, dc *dockerx.DockerClient, cfg *GatewayDeploymentConfig) error {
	containerName := nginxContainerName(cfg.AppName)
	containers, err := dc.ListContainers(ctx, map[string]string{"name": containerName})
	if err != nil {
		return err
	}
	if len(containers) != 1 {
		return fmt.Errorf("unexpected number of containers found: %d", len(containers))
	}
	// asaemble new nginx cfg
	nginxCfg := buildNginxConfig(cfg)
	cfgDirPath, cfgPath, err := createNginxConfigFileInTempDir(nginxCfg)
	if err != nil {
		return fmt.Errorf("failed to create nginx config file: %v", err)
	}
	defer os.RemoveAll(cfgDirPath)
	err = dc.CopyFileToContainer(ctx, containers[0].ID, cfgPath, "/etc/nginx")
	if err != nil {
		return fmt.Errorf("failed to copy nginx config file to container: %v", err)
	}
	err = dc.ExecContainer(ctx, containers[0].ID, []string{"nginx", "-s", "reload"})
	if err != nil && strings.Contains(err.Error(), "no servers are inside upstream") {
		return nil
	}
	return err
}

func (n *nginxGateway) BackendContainersByNetwork(ctx context.Context, dc *dockerx.DockerClient, network string) ([]*dockerx.Container, error) {
	return dc.ListContainers(ctx, map[string]string{"network": network, "label": fmt.Sprintf("gateway=%s", gatewayLabel)})
}

func (n *nginxGateway) Labels() map[string]string {
	return map[string]string{"gateway": gatewayLabel}
}

func (n *nginxGateway) DeployGatewayContainer(ctx context.Context, dc *dockerx.DockerClient, cfg *GatewayDeploymentConfig) error {
	runOpts, err := nginxContainerConfig(cfg)
	if err != nil {
		return err
	}
	// start container
	_, err = startNginxContainer(ctx, dc, runOpts)
	if err != nil {
		return err
	}
	// apply config
	return n.ReloadGatewayContainer(ctx, dc, cfg)
}

func nginxContainerConfig(cfg *GatewayDeploymentConfig) (*dockerx.RunOptions, error) {
	portMapping := make(map[string]string)
	for _, ep := range cfg.ExposedPorts {
		portMapping[ep.Port] = ep.Exposed
	}

	return &dockerx.RunOptions{
		ContainerName: nginxContainerName(cfg.AppName),
		Image:         "nginx",
		Detached:      true,
		Networks:      []string{cfg.Network},
		PortMapping:   portMapping,
	}, nil
}

func createNginxConfigFileInTempDir(nginxCfg string) (dirPath string, cfgPath string, err error) {
	tmpDir, err := os.MkdirTemp("", "nginx-config")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temporary directory for nginx config: %v", err)
	}

	configFilePath := filepath.Join(tmpDir, "nginx.conf")
	if err := os.WriteFile(configFilePath, []byte(nginxCfg), 0644); err != nil {
		defer os.RemoveAll(tmpDir)
		return "", "", fmt.Errorf("failed to write nginx config file: %v", err)
	}
	return tmpDir, configFilePath, nil
}

func nginxContainerName(appName string) string {
	// return fmt.Sprintf("v-nginx-%s-gateway", network)
	return appName
}

func startNginxContainer(ctx context.Context, dc *dockerx.DockerClient, opts *dockerx.RunOptions) (string, error) {
	containerID, err := dc.RunImage(ctx, opts)
	if err != nil {
		return "", fmt.Errorf("failed to start nginx gateway: %v", err)
	}
	return containerID, nil
}

func buildNginxConfig(cfg *GatewayDeploymentConfig) string {
	cfgJson := cfg.Json()
	cfgServerBlock := fmt.Sprintf(`
	server {
		listen %d;
		
		location /config {
			default_type application/json;
			return 200 '%s';
		}
	}
	`, configPort, cfgJson)

	nginxConfig := `
events {}
http {
%s
%s
%s
}
`
	buildNginxUpstream := func(port string) string {
		srvs := make([]string, 0)
		for _, containerName := range cfg.BackendContainerNames {
			srvs = append(srvs, fmt.Sprintf("\t\tserver %s:%s;", containerName, port))
		}
		return fmt.Sprintf(`
	upstream backend-%s {
%s
	}
`, port, strings.Join(srvs, "\n"))
	}

	buildNginxBackendServer := func(port string) string {
		return fmt.Sprintf(`
	server {
		listen %s;
		location / {
			proxy_pass http://backend-%s;
		}
	}
		`, port, port)
	}

	upstreams := []string{}
	backendServers := []string{}
	for _, port := range cfg.Ports {
		upstreams = append(upstreams, buildNginxUpstream(port))
		backendServers = append(backendServers, buildNginxBackendServer(port))
	}

	return fmt.Sprintf(nginxConfig, strings.Join(upstreams, "\n"), strings.Join(backendServers, "\n"), cfgServerBlock)
}
