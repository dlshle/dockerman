package gproxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/dlshle/gommon/logging"
)

type ForwardingPolicy func(ctx context.Context, conn net.Conn, backends []*Backend) *Backend

type Listenr struct {
	protocol string
	port     int
	backends []*Backend
	policy   ForwardingPolicy
}

func NewListener(protocol string, port int, backends []*Backend, policy ForwardingPolicy) *Listenr {
	return &Listenr{
		protocol: protocol,
		port:     port,
		backends: backends,
		policy:   policy,
	}
}

func (l *Listenr) ListenAndServe(ctx context.Context) error {
	port := l.port
	listen, err := net.Listen(l.protocol, fmt.Sprintf("0.0.0.0:%d", port))

	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			logging.GlobalLogger.Infof(ctx, "stopping listener on port %d", port)
			return nil
		default:
			conn, err := listen.Accept()
			if err != nil {
				logging.GlobalLogger.Errorf(ctx, "failed to accept connection: %v, skipping", err)
				continue
			}

			ctx := logging.WrapCtx(ctx, "source", conn.RemoteAddr().String())
			logging.GlobalLogger.Infof(ctx, "connection accepted")
			go func() {
				l.forward(ctx, conn)
			}()
		}
	}
}

func (l *Listenr) forward(ctx context.Context, conn net.Conn) error {
	backend := l.policy(ctx, conn, l.backends)
	logging.GlobalLogger.Infof(ctx, "backend %v has been choosen", backend)
	backendConn, err := net.Dial(l.protocol, backend.Host+":"+strconv.Itoa(backend.Port))
	if err != nil {
		logging.GlobalLogger.Errorf(ctx, "error connecting to backend %v: %v", backend, err)
		return err
	}
	ctx, closeFunc := context.WithCancel(ctx)
	go func() {
		_, err = io.Copy(conn, backendConn)
		closeFunc()
	}()
	go func() {
		_, err = io.Copy(backendConn, conn)
		closeFunc()
	}()
	<-ctx.Done()
	conn.Close()
	backendConn.Close()
	return err
}
