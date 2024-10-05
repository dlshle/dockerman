package server

import (
	"context"
	"time"

	"github.com/dlshle/dockman/proto"
	"github.com/dlshle/gommon/async"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/gts"
	gproto "google.golang.org/protobuf/proto"
)

type TCPProxyServer struct {
	svr          gts.TCPServer
	mgr          *TCPProxyManager
	connReadPool async.AsyncPool
}

func NewTCPProxyServer(port int, mgr *TCPProxyManager) *TCPProxyServer {
	return &TCPProxyServer{
		svr:          gts.NewTCPServer("proxy", "0.0.0.0", port),
		mgr:          mgr,
		connReadPool: async.NewAsyncPool("conn-read-pool", 1024, 512),
	}
}

func (s *TCPProxyServer) Start() error {
	s.initSvr()
	return s.svr.Start()
}

func (s *TCPProxyServer) initSvr() {
	s.svr.OnClientConnected(func(conn gts.Connection) {
		s.handleNewConn(logging.WrapCtx(context.Background(), "ip", conn.Address()), conn)
	})
}

func (s *TCPProxyServer) handleNewConn(ctx context.Context, sourceConn gts.Connection) {
	// wait for the connect request w/ 5 seconds timeout
	sourceIncomingDataChan := make(chan *proto.ProxyMessage, 64)
	sourceConn.OnMessage(func(b []byte) {
		msg := &proto.ProxyMessage{}
		err := gproto.Unmarshal(b, msg)
		if err != nil {
			logging.GlobalLogger.Errorf(ctx, "invalid message received from sourceConn %v: %v, skipping this message", sourceConn, err)
			return
		}
		sourceIncomingDataChan <- msg
	})
	s.connReadPool.Execute(sourceConn.ReadLoop)
	// select once for initial contact
	select {
	case <-time.After(5 * time.Second):
		logging.GlobalLogger.Errorf(ctx, "timed out waiting for connection to be established, closing connection")
		sourceConn.Close()
		return
	case msg := <-sourceIncomingDataChan:
		if msg.Header.Type != proto.ProxyHeader_CONNECT {
			logging.GlobalLogger.Errorf(ctx, "unexpected initial message received from sourceConn %v: %v, closing connection", sourceConn, msg)
			sourceConn.Close()
			return
		}
		if err := s.mgr.handleConnect(ctx, sourceConn, msg, sourceIncomingDataChan); err != nil {
			logging.GlobalLogger.Errorf(ctx, "error handling connect request: %w", err)
			// try to write back response
			sourceConn.Write([]byte("error: " + err.Error()))
			sourceConn.Close()
			return
		}
	}
}
