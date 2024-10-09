package portforward

import (
	"github.com/dlshle/dockman/pkg/dockerx"
	"github.com/dlshle/dockman/pkg/proxy/server"
	"github.com/dlshle/gommon/async"
)

type Portforward struct {
	dc  *dockerx.DockerClient
	svr server.TCPProxyServer
}

func NewPortforwardHandler(port int, dc *dockerx.DockerClient) *Portforward {
	mgr := server.NewTCPProxyManager(async.NewAsyncPool("proxy-processors", 1024, 512), async.NewAsyncPool("proxy-read", 1024, 512), CreatePortforwardBackendResolver(dc))
	return &Portforward{
		dc:  dc,
		svr: *server.NewTCPProxyServer(port, mgr),
	}
}

func (h *Portforward) Start() error {
	return h.svr.Start()
}
