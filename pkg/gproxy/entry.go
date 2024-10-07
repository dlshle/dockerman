package gproxy

import (
	"context"
	"log"
)

func Entry(cfg *Config) (func(), error) {

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
