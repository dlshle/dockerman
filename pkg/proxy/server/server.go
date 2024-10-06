package server

import (
	"context"
	"fmt"

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
	connectMsg, err := s.negotiateNewSourceConn(ctx, sourceConn)
	if err != nil {
		logging.GlobalLogger.Errorf(ctx, "failed to negotiate new source connection: %v", err)
		sourceConn.Close()
		return
	}
	sourceConn.OnMessage(func(b []byte) {
		msg := &proto.ProxyMessage{}
		err := gproto.Unmarshal(b, msg)
		if err != nil {
			logging.GlobalLogger.Errorf(ctx, "invalid message received from sourceConn %v: %v, skipping this message", sourceConn, err)
			return
		}
		sourceIncomingDataChan <- msg
	})
	if err = s.mgr.handleConnect(ctx, sourceConn, connectMsg); err != nil {
		logging.GlobalLogger.Errorf(ctx, "failed to handle connect request from sourceConn %v: %v", sourceConn, err)
		sourceConn.Close()
		return
	}
}

func (s *TCPProxyServer) negotiateNewSourceConn(ctx context.Context, sourceConn gts.Connection) (*proto.ProxyMessage, error) {
	firstStream, err := sourceConn.Read()
	if err != nil {
		return nil, err
	}
	msg := &proto.ProxyMessage{}
	err = gproto.Unmarshal(firstStream, msg)
	if err != nil {
		logging.GlobalLogger.Errorf(ctx, "invalid first stream received from sourceConn %v: %v", sourceConn, err)
		return nil, err
	}
	if msg.Header.Type != proto.ProxyHeader_CONNECT {
		err = fmt.Errorf("unexpected initial message received from sourceConn %v: %v, closing connection", sourceConn, msg)
		logging.GlobalLogger.Error(ctx, err.Error())
		return nil, err
	}
	return msg, nil
}
