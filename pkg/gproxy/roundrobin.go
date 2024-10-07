package gproxy

import (
	"context"
	"net"
	"sync/atomic"
)

func CreateRoundRobinPolicy() ForwardingPolicy {
	var currSelectedBackend int32 = 0
	return func(ctx context.Context, conn net.Conn, backends []*Backend) *Backend {
		currIndex := int(atomic.LoadInt32(&currSelectedBackend))
		if len(backends) <= currIndex {
			currIndex = len(backends)
		}

		nextIndex := (currIndex + 1) % len(backends)
		atomic.StoreInt32(&currSelectedBackend, int32(nextIndex))
		return backends[nextIndex]
	}
}
