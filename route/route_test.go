package route_test

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/gigawattio/go-commons/pkg/web"
	"github.com/gigawattio/go-commons/pkg/web/route"
	"github.com/nbio/hitch"
)

type MyWebService struct {
	*web.WebServer
	NumLoggerInvocations int64
}

func NewMyWebService() *MyWebService {
	mws := &MyWebService{}
	mws.WebServer = web.NewWebServer(web.WebServerOptions{
		Addr:    "127.0.0.1:0",
		Handler: mws.activateRoutes().Handler(),
	})
	return mws
}

func (service *MyWebService) activateRoutes() *hitch.Hitch {
	h := route.Activate(
		[]route.RouteMiddlewareBundle{
			route.RouteMiddlewareBundle{
				Middlewares: []func(http.Handler) http.Handler{
					service.loggerMiddleware,
				},
				RouteData: []route.RouteDatum{
					{"get", "/", func(w http.ResponseWriter, req *http.Request) { fmt.Fprint(w, "hello world") }},
				},
			},
		},
	)
	return h
}

func (service *MyWebService) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		atomic.AddInt64(&service.NumLoggerInvocations, 1)
		next.ServeHTTP(w, req)
	})
}

func TestMiddlewares(t *testing.T) {
	mws := NewMyWebService()
	if err := mws.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := mws.Stop(); err != nil {
			t.Fatal(err)
		}
	}()
	for i := 0; i < 250; i++ {
		response, err := http.Get(mws.BaseUrl())
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode/100 != 2 {
			t.Fatalf("Expected 2xx response status-code but actual status-code=%v", response.StatusCode)
		}
		if expected, actual := int64(i+1), atomic.LoadInt64(&mws.NumLoggerInvocations); actual != expected {
			t.Fatalf("Expected NumLoggerInvocations=%v but actual=%v", expected, actual)
		}
	}
}
