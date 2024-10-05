package client

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/dlshle/gommon/logging"
)

func listen(ctx context.Context, wg *sync.WaitGroup, port int, handler func(context.Context, net.Conn)) error {
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	defer wg.Done()
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			logging.GlobalLogger.Infof(ctx, "stopping port forwarding")
			return nil
		default:
			conn, err := listen.Accept()
			if err != nil {
				logging.GlobalLogger.Errorf(ctx, "failed to accept connection: %w", err)
				continue
			}
			ctx := logging.WrapCtx(ctx, "source", conn.RemoteAddr().String())
			logging.GlobalLogger.Infof(ctx, "connection accepted")
			go handler(ctx, conn)
		}
	}
}
