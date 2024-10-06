package server

import (
	"context"
	"errors"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/dlshle/gommon/async"
	"github.com/dlshle/gommon/logging"

	"github.com/dlshle/dockman/pkg/proxy/common"
	"github.com/dlshle/dockman/proto"
	"github.com/dlshle/gts"
)

type Connection struct {
	ctx         context.Context
	closeFunc   func()
	id          int32
	source      gts.Connection // from dmctl
	destination net.Conn       // to container backend
}

type TCPProxyManager struct {
	hostResolver func(string) (string, error)
	pool         async.AsyncPool
	readPool     async.AsyncPool
	connections  map[int32]*Connection
}

func NewTCPProxyManager(pool async.AsyncPool, readPool async.AsyncPool, hostResolver HostResolver) *TCPProxyManager {
	return &TCPProxyManager{
		pool:         pool,
		readPool:     readPool,
		connections:  make(map[int32]*Connection),
		hostResolver: hostResolver,
	}
}

func (m *TCPProxyManager) disconnectConnByID(ctx context.Context, connID int32) error {
	if conn := m.connections[connID]; conn != nil {
		conn.closeFunc()
		return nil
	}
	return errors.New("connection not found")
}

func (m *TCPProxyManager) handleConnect(ctx context.Context, sourceConn gts.Connection, data *proto.ProxyMessage) error {
	if m.connections[data.Header.ConnectionId] != nil {
		// replace
		logging.GlobalLogger.Infof(ctx, "replace existing connection %d for sourceConn %v", data.Header.ConnectionId, sourceConn)
	}
	connectPayload := data.GetConnectRequest()
	if connectPayload == nil {
		return errors.New("invalid connect request")
	}
	host, err := m.hostResolver(connectPayload.Host)
	if err != nil {
		return errors.New("failed to resolve host: " + err.Error())
	}
	destConn, err := net.Dial("tcp", host+":"+strconv.Itoa(int(connectPayload.Port)))
	if err != nil {
		return err
	}

	// reply ack to source once backend connection successfully established
	if err = sourceConn.Write([]byte("ok")); err != nil {
		logging.GlobalLogger.Errorf(ctx, "error writing ack to sourceConn %v: %v", sourceConn, err)
		destConn.Close()
		return err
	}

	connCtx, closeFunc := context.WithCancel(ctx)

	destConn.SetDeadline(time.Time{})

	conn := &Connection{
		ctx:         connCtx,
		closeFunc:   closeFunc,
		id:          data.Header.ConnectionId,
		source:      sourceConn,
		destination: destConn,
	}
	m.connections[data.Header.ConnectionId] = conn
	m.pool.Execute(func() { m.connectionProcessLoop(conn) })
	return nil
}

func (m *TCPProxyManager) connectionProcessLoop(conn *Connection) {
	closeFunc := conn.closeFunc
	connCtx := conn.ctx
	sourceConn := conn.source
	backendConn := conn.destination

	proxySourceConn := common.NewProxyConn(conn.id, sourceConn)

	// backend -> source
	m.readPool.Execute(func() {
		_, err := io.Copy(proxySourceConn, backendConn)
		if err != nil {
			logging.GlobalLogger.Errorf(connCtx, "error occurred while copying data from backend to source: %v, closing connection", err)
			closeFunc()
		}
	})

	// source -> backend
	m.readPool.Execute(func() {
		_, err := proxySourceConn.WriteTo(backendConn)
		if err != nil {
			logging.GlobalLogger.Errorf(connCtx, "error occurred while copying data from source to backend: %v, closing connection", err)
			closeFunc()
		}
	})

	<-connCtx.Done()
	sourceConn.Close()
	backendConn.Close()
	delete(m.connections, conn.id)
}
