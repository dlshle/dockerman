package gproxy

import "errors"

func CreatePolicyFactories() func(string) (func() ForwardingPolicy, error) {
	factories := map[string]func() ForwardingPolicy{
		"roundrobin": CreateRoundRobinPolicy,
	}
	return func(s string) (func() ForwardingPolicy, error) {
		policyFactory, exists := factories[s]
		if !exists {
			return nil, errors.New("unknown policy: " + s)
		}
		return policyFactory, nil
	}
}
