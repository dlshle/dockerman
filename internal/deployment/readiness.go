package deployment

import (
	"context"
	"fmt"
	"time"

	"github.com/dlshle/dockman/internal/config"
	"github.com/dlshle/gommon/http"
	"github.com/dlshle/gommon/logging"
)

func (d *Deployment) checkRediness(ctx context.Context, appCfg *config.AppConfig, replicaID int) error {
	// OK, another problem, if port isn't exposed on the container, I can't directly send HTTP rqeuest to check rediness
	if appCfg.ReadinessCheck == nil {
		return nil
	}
	container, err := d.containerByReplicaID(ctx, appCfg, replicaID)
	containerName := container.Names[0]
	if err != nil {
		return err
	}
	network := networkNameByAppConfig(appCfg)
	containerIP, networkExists := container.IPAddresses[network]
	if !networkExists {
		return fmt.Errorf("container %s does not have network %s", containerName, network)
	}
	logging.GlobalLogger.Infof(ctx, "checking rediness of container %s with ip %s", containerName, containerIP)
	backendClient := d.backendHTTPClient(appCfg, replicaID)
	url := fmt.Sprintf("http://%s:%d%s", containerIP, appCfg.ReadinessCheck.Port, appCfg.ReadinessCheck.Path)
	method := appCfg.ReadinessCheck.Method
	checkInterval := time.Second // 1 second check interval by default
	if appCfg.ReadinessCheck.CheckIntervalSeconds != nil {
		checkInterval = time.Duration(*appCfg.ReadinessCheck.CheckIntervalSeconds) * time.Second
	}
	maxRetries := 30
	if appCfg.ReadinessCheck.HealthCheckFailureThreshold != nil {
		maxRetries = *appCfg.ReadinessCheck.HealthCheckFailureThreshold
	}
	return tryRedinessCheckAtFixedInterval(ctx, backendClient, method, url, checkInterval, maxRetries)
}

func tryRedinessCheckAtFixedInterval(ctx context.Context, client http.HTTPClient, method string, url string, interval time.Duration, maxRetries int) (err error) {
	for i := 0; i < maxRetries; i++ {
		// TODO what it should do is to try readiness check in async so that next retry can happen in time
		if err = doRedinessCheck(ctx, client, method, url, interval); err == nil {
			return
		}
		logging.GlobalLogger.Warnf(ctx, "rediness check failed, retrying in %s", interval)
		time.Sleep(interval)
	}
	return
}

func doRedinessCheck(ctx context.Context, client http.HTTPClient, method string, url string, timeout time.Duration) error {
	resp := client.Request(http.NewRequestBuilder().URL(url).Method(method).Timeout(timeout).Build())
	if resp.Success {
		return nil
	}
	return fmt.Errorf("rediness check failed with status code %d (%s)", resp.Code, resp.Body)
}
