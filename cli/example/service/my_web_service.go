package service

import (
	"fmt"
	"net/http"

	"github.com/gigawattio/go-commons/pkg/web"
	"github.com/gigawattio/go-commons/pkg/web/route"

	cliv2 "gopkg.in/urfave/cli.v2"
)

type MyWebService struct {
	*web.WebServer
}

func (mws *MyWebService) handler() http.Handler {
	h := route.Activate(
		[]route.RouteMiddlewareBundle{
			route.RouteMiddlewareBundle{
				RouteData: []route.RouteDatum{
					{"get", "/", func(w http.ResponseWriter, r *http.Request) {
						fmt.Fprint(w, "hello world")
					}},
				},
			},
		},
	).Handler()
	return h
}

func (mws *MyWebService) Start(ctx *cliv2.Context) error {
	if mws.WebServer == nil {
		options := web.WebServerOptions{
			Addr:    ctx.String("bind"),
			Handler: mws.handler(),
		}
		mws.WebServer = web.NewWebServer(options)
	}
	if err := mws.WebServer.Start(); err != nil {
		return err
	}
	return nil
}

func (mws *MyWebService) Stop() error {
	if err := mws.WebServer.Stop(); err != nil {
		return err
	}
	return nil
}
