package gproxy

import (
	"context"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

func entry(cfgPath string) (func(), error) {
	cfgData, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err = yaml.UnmarshalStrict(cfgData, cfg); err != nil {
		return nil, err
	}
	policyFactories := CreatePolicyFactories()

	ctx, cancelFunc := context.WithCancel(context.Background())
	for _, upstream := range cfg.Upstreams {
		policyFactory, err := policyFactories(upstream.Policy)
		if err != nil {
			cancelFunc()
			return nil, err
		}
		l := NewListener(upstream.Protocol, upstream.Port, upstream.Backends, policyFactory())
		go func() {
			err := l.ListenAndServe(ctx)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
	return cancelFunc, nil
}
