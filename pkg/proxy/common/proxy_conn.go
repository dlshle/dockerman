package common

import (
	"errors"
	"io"

	"github.com/dlshle/dockman/proto"
	"github.com/dlshle/gts"
	gproto "google.golang.org/protobuf/proto"
)

type ProxyConn struct {
	connID int32
	conn   gts.Connection
}

func NewProxyConn(connID int32, conn gts.Connection) *ProxyConn {
	return &ProxyConn{connID: connID, conn: conn}
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
		return 0, err
	}
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
