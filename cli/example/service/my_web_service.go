package service

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gigawattio/go-commons/pkg/web"
	"github.com/gigawattio/go-commons/pkg/web/route"
)

type MyWebService struct {
	*web.WebServer
}

func New(bind string) *MyWebService {
	service := &MyWebService{}
	options := web.WebServerOptions{
		Addr:    bind,
		Handler: service.activateRoutes(),
	}
	service.WebServer = web.NewWebServer(options)
	return service
}

func (mws *MyWebService) activateRoutes() http.Handler {
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

func (service *MyWebService) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(os.Stderr, "method=%s url=%s remoteAddr=%s referer=%s\n", req.Method, req.URL.String(), req.RemoteAddr, req.Referer())
		next.ServeHTTP(w, req)
	})
}
