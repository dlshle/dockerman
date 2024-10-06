package client

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/dlshle/gommon/logging"
)

func PortForward(ctx context.Context, wg *sync.WaitGroup, proxySvrHost string, localPort int, remote *Remote) error {
	ctx = logging.WrapCtx(ctx, "portMapping", fmt.Sprintf("%d <-> %s:%d", localPort, remote.Host, remote.Port))
	logging.GlobalLogger.Infof(ctx, "starting port forwarding")
	return listen(ctx, wg, localPort, func(ctx context.Context, conn net.Conn) {
		p := NewProxy(ctx, conn, remote)
		if err := p.init(proxySvrHost); err != nil {
			logging.GlobalLogger.Errorf(ctx, "failed to create proxy: %v", err)
		}
		conn.Close()
	})
}
