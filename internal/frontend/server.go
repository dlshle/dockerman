package frontend

import (
	"context"
	"fmt"

	"github.com/dlshle/aghs/server"
	"github.com/dlshle/dockman/internal/config"
	"github.com/dlshle/dockman/internal/handler"
	"gopkg.in/yaml.v2"
)

func ServeHTTP(port int, dmHandler *handler.DockmanHandler) error {
	svc, err := server.NewServiceBuilder().Id("dockman").
		WithRouteHandlers(server.PathHandlerBuilder("/deploy").
			Post(server.NewCHandlerBuilder[*config.AppConfig]().Unmarshaller(func(b []byte) (*config.AppConfig, error) {
				cfg := &config.AppConfig{}
				err := yaml.Unmarshal(b, cfg)
				return cfg, err
			}).OnRequest(func(c server.CHandle[*config.AppConfig]) server.Response {
				err := dmHandler.Deploy(context.Background(), c.Data())
				if err != nil {
					return server.NewPlainTextResponse(500, err)
				}
				return server.NewPlainTextResponse(200, "ok")
			}).MustBuild().HandleRequest).
			Build()).
		Build()
	if err != nil {
		return err
	}
	svr, err := server.NewBuilder().Address(fmt.Sprintf("0.0.0.0:%d", port)).
		WithService(svc).
		Build()
	if err != nil {
		return err
	}
	return svr.Start()
}
