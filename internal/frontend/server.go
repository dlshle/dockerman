package frontend

import (
	"context"
	"errors"
	"fmt"

	"github.com/dlshle/aghs/server"
	"github.com/dlshle/dockman/internal/config"
	"github.com/dlshle/dockman/internal/handler"
	"github.com/dlshle/gommon/logging"
	"gopkg.in/yaml.v2"
)

func ServeHTTP(port int, dmHandler *handler.DockmanHandler) error {
	svc, err := server.NewServiceBuilder().Id("dockman").
		WithRouteHandlers(server.PathHandlerBuilder("/ping").Get(server.NewCHandlerBuilder[any]().OnRequest(func(c server.CHandle[any]) server.Response {
			return server.NewPlainTextResponse(200, "ok")
		}).MustBuild().HandleRequest)).
		WithRouteHandlers(server.PathHandlerBuilder("/deploy").
			Post(server.NewCHandlerBuilder[*config.AppConfig]().RequireBody().Unmarshaller(func(b []byte) (*config.AppConfig, error) {
				if len(b) == 0 {
					return nil, errors.New("invalid argument: empty body")
				}
				cfg := &config.AppConfig{}
				err := yaml.UnmarshalStrict(b, cfg)
				return cfg, err
			}).OnRequest(func(c server.CHandle[*config.AppConfig]) server.Response {
				err := dmHandler.Deploy(logging.WrapCtx(context.Background(), "traceId", c.Request().Id()), c.Data())
				if err != nil {
					return server.NewPlainTextResponse(500, err.Error())
				}
				return server.NewPlainTextResponse(200, "ok")
			}).MustBuild().HandleRequest).
			Build()).
		WithRouteHandlers(server.PathHandlerBuilder("/rollout").
			Post(server.NewCHandlerBuilder[*config.AppConfig]().RequireBody().Unmarshaller(func(b []byte) (*config.AppConfig, error) {
				if len(b) == 0 {
					return nil, errors.New("invalid argument: empty body")
				}
				cfg := &config.AppConfig{}
				err := yaml.UnmarshalStrict(b, cfg)
				return cfg, err
			}).OnRequest(func(c server.CHandle[*config.AppConfig]) server.Response {
				err := dmHandler.Rollout(logging.WrapCtx(context.Background(), "traceId", c.Request().Id()), c.Data())
				if err != nil {
					return server.NewPlainTextResponse(500, err.Error())
				}
				return server.NewPlainTextResponse(200, "ok")
			}).MustBuild().HandleRequest).
			Build()).
		WithRouteHandlers(server.PathHandlerBuilder("/deployments").
			Get(server.NewCHandlerBuilder[any]().OnRequest(func(c server.CHandle[any]) server.Response {
				deployments, err := dmHandler.ListDeployments(logging.WrapCtx(context.Background(), "traceId", c.Request().Id()))
				if err != nil {
					return server.NewPlainTextResponse(500, err.Error())
				}
				return server.NewResponse(200, deployments)
			}).MustBuild().HandleRequest)).
		WithRouteHandlers(server.PathHandlerBuilder("/deployment/:id").
			Delete(server.NewCHandlerBuilder[any]().AddRequiredPathParam("id").OnRequest(func(c server.CHandle[any]) server.Response {
				err := dmHandler.Delete(logging.WrapCtx(context.Background(), "traceId", c.Request().Id()), c.PathParam("id"))
				if err != nil {
					return server.NewPlainTextResponse(500, err.Error())
				}
				return server.NewPlainTextResponse(200, "ok")
			}).MustBuild().HandleRequest).
			Get(server.NewCHandlerBuilder[any]().AddRequiredPathParam("id").OnRequest(func(c server.CHandle[any]) server.Response {
				info, err := dmHandler.InfoDeployment(logging.WrapCtx(context.Background(), "traceId", c.Request().Id()), c.PathParam("id"))
				if err != nil {
					return server.NewPlainTextResponse(500, err.Error())
				}
				return server.NewResponse(200, info)
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
