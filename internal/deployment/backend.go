package deployment

import (
	"context"

	"github.com/dlshle/dockman/internal/config"
	"github.com/dlshle/dockman/pkg/dockerx"
	"github.com/dlshle/gommon/http"
)

// returns containerID + error
func (d *Deployment) deployReplica(ctx context.Context, appCfg *config.AppConfig, replicaID int) (string, error) {
	runOpts := &dockerx.RunOptions{
		ContainerName: containerNameByAppConfig(appCfg, replicaID),
		Detached:      true,
		Image:         appCfg.Image,
		Networks:      []string{networkNameByAppConfig(appCfg)},
	}
	return d.docker.RunImage(ctx, runOpts)
}

func (d *Deployment) backendHTTPClient(appCfg *config.AppConfig, replicaID int) http.HTTPClient {
	containerName := containerNameByAppConfig(appCfg, replicaID)
	containerClient := d.backendClients[containerName]
	if containerClient == nil {
		containerClient = http.NewBuilder().Id(containerName).MaxConcurrentRequests(32).MaxConnsPerHost(8).Build()
		d.backendClients[containerName] = containerClient
	}
	return containerClient
}
