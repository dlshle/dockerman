package dmdaemon

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/dlshle/dockman/internal/config"
	"github.com/dlshle/gommon/http"
	"gopkg.in/yaml.v2"
)

type DmDaemonCli struct {
	host    string
	httpCli http.HTTPClient
}

func NewCli(host string) (*DmDaemonCli, error) {
	httpCli := http.NewHTTPClient(5, 128, 10)
	cli := &DmDaemonCli{
		host:    host,
		httpCli: httpCli,
	}
	return cli, cli.Ping(context.Background())
}

func (c *DmDaemonCli) ListDeployments(ctx context.Context) ([]string, error) {
	req := http.NewRequestBuilder().Context(ctx).Method("GET").URL(c.host + "/deployments").Build()
	resp := c.httpCli.Request(req)
	if !resp.Success {
		return nil, fmt.Errorf("failed to list deployments: %s", resp.Body)
	}
	res := make([]string, 0)
	err := json.Unmarshal([]byte(resp.Body), &res)
	return res, err
}

func (c *DmDaemonCli) Rollout(ctx context.Context, appCfg *config.AppConfig) error {
	marshalled, err := yaml.Marshal(appCfg)
	if err != nil {
		return err
	}
	req := http.NewRequestBuilder().Method("POST").URL(c.host + "/rollout").Body(io.NopCloser(bytes.NewReader(marshalled))).Timeout(time.Minute * 5).Build()
	resp := c.httpCli.Request(req)
	if !resp.Success {
		return fmt.Errorf("failed to rollout: %s", resp.Body)
	}
	return nil
}

func (c *DmDaemonCli) GetDeployment(ctx context.Context, appName string) (*config.AppConfig, error) {
	req := http.NewRequestBuilder().Method("GET").URL(c.host + "/deployments/" + appName).Build()
	resp := c.httpCli.Request(req)
	if !resp.Success {
		return nil, fmt.Errorf("failed to get deployment %s: %s", appName, resp.Body)
	}
	appCfg := new(config.AppConfig)
	err := yaml.Unmarshal([]byte(resp.Body), &appCfg)
	return appCfg, err
}

func (c *DmDaemonCli) DeleteDeployment(ctx context.Context, appName string) error {
	req := http.NewRequestBuilder().Method("DELETE").URL(c.host + "/deployments/" + appName).Build()
	resp := c.httpCli.Request(req)
	if !resp.Success {
		return fmt.Errorf("failed to delete deployment %s: %s", appName, resp.Body)
	}
	return nil
}

func (c *DmDaemonCli) Ping(ctx context.Context) error {
	req := http.NewRequestBuilder().Method("GET").URL(c.host + "/ping").Build()
	resp := c.httpCli.Request(req)
	if !resp.Success {
		return fmt.Errorf("ping to %s failed: %s", c.host+"/ping", resp.Body)
	}
	return nil
}
