package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/dlshle/gommon/async"
	"github.com/dlshle/gommon/logging"

	"github.com/dlshle/dockman/proto"
	"github.com/dlshle/gts"
	gproto "google.golang.org/protobuf/proto"
)

type Connection struct {
	ctx         context.Context
	closeFunc   func()
	id          int32
	source      gts.Connection // from dmctl
	destination net.Conn       // to container backend
	fromSource  chan *proto.ProxyMessage
	fromDest    chan []byte
	toSource    chan []byte
	toDest      chan []byte
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

func (m *TCPProxyManager) handleConnect(ctx context.Context, sourceConn gts.Connection, data *proto.ProxyMessage, incomingDataChan chan *proto.ProxyMessage) error {
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

	fromDestToSourceDataChan := m.prepareBackendConn(connCtx, closeFunc, destConn)

	conn := &Connection{
		ctx:         ctx,
		closeFunc:   nil,
		id:          data.Header.ConnectionId,
		source:      sourceConn,
		destination: destConn,
		fromSource:  incomingDataChan,
		fromDest:    fromDestToSourceDataChan,
		toSource:    make(chan []byte, 64),
		toDest:      make(chan []byte, 64),
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

	// read process loops
	m.pool.Execute(func() {
		for {
			select {
			case <-connCtx.Done():
				return
			case destToSourceData := <-conn.fromDest:
				// for source conn, we need to write formatted data
				msg := &proto.ProxyMessage{
					Header: &proto.ProxyHeader{
						ConnectionId: conn.id,
					},
					Data: &proto.ProxyMessage_Payload{
						Payload: destToSourceData,
					},
				}
				data, err := gproto.Marshal(msg)
				if err != nil {
					logging.GlobalLogger.Errorf(connCtx, "error marshalling message %v, skipping", err)
					continue
				}
				logging.GlobalLogger.Debugf(connCtx, "received data from backend: %v", msg)
				select {
				case <-time.After(5 * time.Second):
					logging.GlobalLogger.Errorf(connCtx, "timed out waiting for message to be sent, closing connection")
					closeFunc()
				case conn.toSource <- data:
					logging.GlobalLogger.Debugf(connCtx, "wrote data to source: %v", msg)
				}
			case sourceToDestMsg := <-conn.fromSource:
				if err := m.handlePayload(conn, sourceToDestMsg); err != nil {
					logging.GlobalLogger.Errorf(connCtx, "Failed to handle payload: %v", err)
					closeFunc()
				}
			}
		}
	})

	// write process loops
	for {
		select {
		case <-connCtx.Done():
			logging.GlobalLogger.Errorf(connCtx, "Connection closed, removing conn from proxy")
			sourceConn.Close()
			backendConn.Close()
			delete(m.connections, conn.id)
			close(conn.fromDest)
			close(conn.fromSource)
			return
		case writeToSourceData := <-conn.toSource:
			if err := sourceConn.Write(writeToSourceData); err != nil {
				logging.GlobalLogger.Errorf(connCtx, "Failed to write to source conn: %v, closing connection", err)
				closeFunc()
			} else {
				logging.GlobalLogger.Infof(connCtx, "wrote %s to source conn", string(writeToSourceData))
			}
		case writeToDestData := <-conn.toDest:
			// TODO add exponential backoff retry
			_, err := backendConn.Write(writeToDestData)
			if err != nil {
				logging.GlobalLogger.Errorf(connCtx, "Failed to write to backend conn: %v, closing connection", err)
				closeFunc()
			} else {
				logging.GlobalLogger.Infof(connCtx, "wrote %s to backend conn", string(writeToDestData))
			}
		}
	}
}

// TODO maybe we don't need this since we are actively reading from the source, and we can use handlePayload to proxy write to dest data
func (m *TCPProxyManager) sourceReadLoop(connCtx context.Context, closeFunc func(), sourceConn gts.Connection, dataChan chan []byte) {
	for {
		select {
		case <-connCtx.Done():
			return
		default:
			if data, err := sourceConn.Read(); err == nil {
				dataChan <- data
			} else {
				logging.GlobalLogger.Errorf(connCtx, "error reading from backend connection: %v", err)
				sourceConn.Close()
				closeFunc()
			}
		}
	}
}

// we need this initiated from the process loop
func (m *TCPProxyManager) backendReadLoop(connCtx context.Context, closeFunc func(), backendConn net.Conn, dataChan chan []byte) {
	for {
		select {
		case <-connCtx.Done():
			return
		default:
			backendConn.Read(nil)
			if data, err := io.ReadAll(backendConn); err == nil {
				dataChan <- data
			} else {
				logging.GlobalLogger.Errorf(connCtx, "error reading from backend connection: %v", err)
				backendConn.Close()
				closeFunc()
			}
		}
	}
}

func (m *TCPProxyManager) handlePayload(conn *Connection, data *proto.ProxyMessage) error {
	logging.GlobalLogger.Infof(conn.ctx, "received msg from source: %v", data)
	if payload := data.GetPayload(); len(payload) > 0 {
		// handle payload
		select {
		case <-time.After(5 * time.Second):
			return fmt.Errorf("source to destination data write timeout(5 seconds)")
		case conn.toDest <- payload:
			// good
			return nil
		}
	} else if disconnectReq := data.GetDisconnectRequest(); disconnectReq != nil {
		// handle disconnect
		conn.closeFunc()
		return nil
	}
	return fmt.Errorf("unknown data type")
}

func (m *TCPProxyManager) prepareBackendConn(ctx context.Context, closeFunc func(), conn net.Conn) chan []byte {
	dataChan := make(chan []byte, 64)
	conn.SetDeadline(time.Time{})
	m.readPool.Execute(func() { m.backendReadLoop(ctx, closeFunc, conn, dataChan) })
	return dataChan
}
