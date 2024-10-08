package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/dlshle/dockman/pkg/proxy/common"
	"github.com/dlshle/dockman/pkg/randx"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/gts"
)

type Proxy struct {
	id         int32
	ctx        context.Context
	closeFunc  func()
	remote     *Remote
	sourceConn net.Conn
	destConn   gts.Connection
}

func NewProxy(ctx context.Context, sourceConn net.Conn, remote *Remote) *Proxy {
	ctx = logging.WrapCtx(ctx, "remote", fmt.Sprintf("%s:%d", remote.Host, remote.Port))
	ctx, closeFunc := context.WithCancel(ctx)
	p := &Proxy{
		ctx:        ctx,
		closeFunc:  closeFunc,
		remote:     remote,
		sourceConn: sourceConn,
	}
	return p
}

func (p *Proxy) init(proxySvrHost string) error {
	conn := p.sourceConn
	// prepare local tcp conn
	conn.SetDeadline(time.Time{})

	// establish conn with backend
	backendConn, err := net.Dial("tcp", proxySvrHost)
	if err != nil {
		return err
	}
	p.destConn = gts.NewTCPConnection(backendConn)

	// we need backend writer since we are writing data to backend with special format
	var connID int32 = randx.Int32()
	backendProxy := common.NewProxyConn(connID, p.destConn)

	// negotiate with backend
	if err = p.backendConnectNegotiation(backendProxy); err != nil {
		logging.GlobalLogger.Errorf(context.Background(), "failed to negotiate with backend: %v", err)
		return err
	}
	p.startProxyLoop(backendProxy)
	return nil
}

func (p *Proxy) startProxyLoop(backendProxy *common.ProxyConn) {
	ctx := p.ctx
	// from source -> proxy
	go func() {
		_, err := io.Copy(backendProxy, p.sourceConn)
		if err != nil {
			logging.GlobalLogger.Errorf(ctx, "error occurred while copying data from client to destination: %v", err)
			p.closeFunc()
		}
	}()
	// proxy -> source
	go func() {
		_, err := backendProxy.WriteTo(p.sourceConn)
		if err != nil {
			logging.GlobalLogger.Errorf(ctx, "error occurred while copying data from destination to client: %v", err)
			p.closeFunc()
		}
	}()

	// wait till connection is closed
	<-ctx.Done()
	logging.GlobalLogger.Infof(ctx, "connection closed")
	p.sourceConn.Close()
	p.destConn.Close()
}

func (p *Proxy) backendConnectNegotiation(backendProxy *common.ProxyConn) error {
	if err := backendProxy.WriteConnect(p.id, p.remote.Host, p.remote.Port); err != nil {
		return err
	}

	reply, err := p.destConn.Read()
	if err != nil {
		return err
	}
	if string(reply) != "ok" {
		return errors.New("unexpected reply from backend")
	}
	return nil
}
