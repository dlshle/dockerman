package main

import (
	"fmt"
	"time"

	"github.com/dlshle/dockman/internal/probe"
	"github.com/dlshle/gommon/http"
)

var (
	client http.HTTPClient = http.NewBuilder().Id("probe").MaxConcurrentRequests(512).MaxConnsPerHost(8).Build()
)

func probeHTTP(cfg *probe.HTTPProbeConfig) error {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5
	}
	requestBuilder := http.NewRequestBuilder().Method(cfg.Method).URL(cfg.URL).Timeout(time.Duration(cfg.Timeout * int(time.Second)))
	if len(cfg.Headers) > 0 {
		headerMaker := http.NewHeaderMaker()
		for _, h := range cfg.Headers {
			headerMaker.Set(h.Key, h.Value)
		}
		requestBuilder.Header(headerMaker.Make())
	}
	resp := client.Request(requestBuilder.Build())
	if cfg.ExpectedStatus != nil && *cfg.ExpectedStatus != resp.Code {
		return fmt.Errorf("expected status code %d, got %d", *cfg.ExpectedStatus, resp.Code)
	}
	if !resp.Success {
		return fmt.Errorf("request failed with status code %d", resp.Code)
	}
	return nil
}
