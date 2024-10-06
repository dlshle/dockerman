package common

import (
	"context"
	"errors"
	"io"

	"github.com/dlshle/dockman/proto"
	"github.com/dlshle/gommon/logging"
	"github.com/dlshle/gts"
	gproto "google.golang.org/protobuf/proto"
)

type ProxyConn struct {
	connID          int32
	conn            gts.Connection
	readInterceptor func(msg *proto.ProxyMessage) error
}

func NewProxyConn(connID int32, conn gts.Connection) *ProxyConn {
	return &ProxyConn{connID: connID, conn: conn}
}

func NewProxyConnWithReadInterceptor(connID int32, conn gts.Connection, interceptor func(msg *proto.ProxyMessage) error) *ProxyConn {
	return &ProxyConn{connID: connID, conn: conn, readInterceptor: interceptor}
}

func (c *ProxyConn) WriteTo(writer io.Writer) (int64, error) {
	for {
		payload, err := c.read()
		if err != nil {
			return 0, err
		}
		if _, err = writer.Write(payload); err != nil {
			return 0, err
		}
	}
}

func (c *ProxyConn) read() ([]byte, error) {
	data, err := c.conn.Read()
	if err != nil {
		return nil, err
	}
	msg := &proto.ProxyMessage{}
	if err = gproto.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	if c.readInterceptor != nil {
		if err = c.readInterceptor(msg); err != nil {
			return nil, err
		}
	}
	if len(msg.GetPayload()) > 0 {
		logging.GlobalLogger.Debugf(context.Background(), "read message %v", msg)
	}
	if msg.GetDisconnectRequest() != nil {
		return nil, io.EOF
	}
	if msg.GetPayload() == nil {
		return nil, errors.New("invalid message received")
	}
	return msg.GetPayload(), nil
}

func (c *ProxyConn) Write(data []byte) (int, error) {
	msg := &proto.ProxyMessage{
		Header: &proto.ProxyHeader{
			ConnectionId: c.connID,
			Type:         proto.ProxyHeader_DATA,
		},
		Data: &proto.ProxyMessage_Payload{
			Payload: data,
		},
	}
	marshalled, err := gproto.Marshal(msg)
	if err != nil {
		return 0, err
	}
	if err = c.conn.Write(marshalled); err != nil {
		logging.GlobalLogger.Errorf(context.Background(), "error writing message: %v due to %v", msg, err)
		return 0, err
	}
	logging.GlobalLogger.Debugf(context.Background(), "wrote message: %v", msg)
	return len(data), nil
}

func (c *ProxyConn) WriteConnect(connID int32, backendHost string, backendPort int32) error {
	msg := &proto.ProxyMessage{
		Header: &proto.ProxyHeader{
			ConnectionId: c.connID,
			Type:         proto.ProxyHeader_CONNECT,
		},
		Data: &proto.ProxyMessage_ConnectRequest{
			ConnectRequest: &proto.ConnectRequest{
				Host: backendHost,
				Port: backendPort,
			},
		},
	}
	marshalled, err := gproto.Marshal(msg)
	if err != nil {
		return err
	}
	return c.conn.Write(marshalled)
}
