package gproxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"slices"
	"strconv"
	"sync"

	"github.com/dlshle/gommon/logging"
	slicesx "github.com/dlshle/gommon/slices"
)

type ForwardingPolicy func(ctx context.Context, conn net.Conn, backends []*Backend) *Backend

type Listener struct {
	ctx       context.Context
	closeFunc func()
	mutex     *sync.Mutex
	conns     map[string][]net.Conn // backend addr: net.Conn
	protocol  string
	port      int
	backends  []*Backend
	policy    ForwardingPolicy
}

func NewListener(ctx context.Context, protocol string, port int, backends []*Backend, policy ForwardingPolicy) *Listener {
	ctx, closeFunc := context.WithCancel(ctx)
	return &Listener{
		ctx:       ctx,
		closeFunc: closeFunc,
		conns:     make(map[string][]net.Conn),
		mutex:     new(sync.Mutex),
		protocol:  protocol,
		port:      port,
		backends:  backends,
		policy:    policy,
	}
}

func (l *Listener) UpdateBackends(backends []*Backend) error {
	currBackends := slicesx.ToMap(l.safeBackends(), func(backend *Backend) (string, *Backend) {
		return backend.Addr(), backend
	})
	newBackends := slicesx.ToMap(backends, func(backend *Backend) (string, *Backend) {
		return backend.Addr(), backend
	})
	toDeleteSet := make(map[string]bool)
	for addr := range currBackends {
		if newBackends[addr] == nil {
			toDeleteSet[addr] = true
		}
	}

	// close backend conns
	for addr := range toDeleteSet {
		for _, conn := range l.getBackendConns(addr) {
			conn.Close()
		}
	}

	// delete toDelete backends
	l.mutex.Lock()
	slices.DeleteFunc(l.backends, func(backend *Backend) bool {
		return toDeleteSet[backend.Addr()]
	})
	l.mutex.Unlock()

	// close connections to the toDelete backends
	for addr := range toDeleteSet {
		l.getBackendConns(addr)
	}

	// update current cfg
	l.mutex.Lock()
	l.backends = backends
	l.mutex.Unlock()

	return nil
}

func (l *Listener) ListenAndServe() error {
	ctx := l.ctx
	port := l.port
	listen, err := net.Listen(l.protocol, fmt.Sprintf("0.0.0.0:%d", port))
	logging.GlobalLogger.Infof(ctx, "listener on port %d", port)
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

func (l *Listener) safeBackends() []*Backend {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.backends
}

func (l *Listener) getBackendConns(backendAddr string) []net.Conn {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.conns[backendAddr]
}

func (l *Listener) addConn(backendAddr string, conn net.Conn) {
	l.mutex.Lock()
	backendConns := l.conns[backendAddr]
	if backendConns == nil {
		backendConns = make([]net.Conn, 0)
	}
	backendConns = append(backendConns, conn)
	l.conns[backendAddr] = backendConns
	l.mutex.Unlock()
}

func (l *Listener) deleteConn(backendAddr string, conn net.Conn) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	backendConns := l.conns[backendAddr]
	if backendConns == nil {
		return
	}
	backendConns = slices.DeleteFunc(backendConns, func(c net.Conn) bool { return c == conn })
	if len(backendConns) == 0 {
		delete(l.conns, backendAddr)
		return
	}
	l.conns[backendAddr] = backendConns
}

func (l *Listener) forward(ctx context.Context, conn net.Conn) error {
	defer conn.Close()
	backend := l.policy(ctx, conn, l.safeBackends())
	logging.GlobalLogger.Infof(ctx, "backend %v has been choosen", backend)
	backendConn, err := net.Dial(l.protocol, backend.Host+":"+strconv.Itoa(backend.Port))
	if err != nil {
		logging.GlobalLogger.Errorf(ctx, "error connecting to backend %v: %v", backend, err)
		return err
	}
	defer backendConn.Close()
	ctx, closeFunc := context.WithCancel(ctx)
	l.addConn(backend.Addr(), conn)
	defer l.deleteConn(backend.Addr(), conn)
	go func() {
		_, err = io.Copy(conn, backendConn)
		closeFunc()
	}()
	_, err = io.Copy(backendConn, conn)
	closeFunc()
	<-ctx.Done()
	return err
}
