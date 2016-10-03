package generics

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gigawattio/go-commons/pkg/web"
	"github.com/gigawattio/go-commons/pkg/web/route"
	"github.com/nbio/hitch"
	"github.com/parnurzeal/gorequest"
)

func genRoutes() *hitch.Hitch {
	index := func(w http.ResponseWriter, req *http.Request) {
		web.RespondWithHtml(w, 200, `<html><head><title>hello world</title></head><body>hello world</body></html>`)
	}
	object := func(w http.ResponseWriter, req *http.Request) {
		GenericObjectEndpoint(w, req, func() (interface{}, error) {
			if req.Method != http.MethodPost {
				return nil, errors.New("Only POST requests are permitted to the `object' endpoint")
			}
			return struct{ Success bool }{Success: true}, nil
		})
	}
	objects := func(w http.ResponseWriter, req *http.Request) {
		GenericObjectsEndpoint(w, req, func(limit int64, offset int64) (interface{}, int, error) {
			if req.Method != http.MethodPost {
				return nil, 0, errors.New("Only POST requests are permitted to the `objects' endpoint")
			}
			return []string{"a", "b", "c", "d"}, 4, nil
		})
	}
	routes := []route.RouteMiddlewareBundle{
		route.RouteMiddlewareBundle{
			RouteData: []route.RouteDatum{
				{"get", "/", index},
				{"post", "/v1/object", object},
				{"post", "/v1/objects", objects},
			},
		},
	}
	h := route.Activate(routes)
	return h
}

func TestGenerics(t *testing.T) {
	options := web.WebServerOptions{
		Addr:    "127.0.0.1:0",
		Handler: genRoutes().Handler(),
	}
	webServer := web.NewWebServer(options)

	if err := webServer.Start(); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := webServer.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	baseUrl := fmt.Sprintf("http://%v", webServer.Addr())

	{
		response, body, errs := gorequest.New().Get(baseUrl + "/").End()
		if len(errs) > 0 {
			t.Fatalf("Error(s) getting /: %+v", errs)
		}
		if response.StatusCode/100 != 2 {
			t.Fatalf("Expected 2xx status-code but actual=%v; body=%v", response.StatusCode, body)
		}
	}

	{
		response, body, errs := gorequest.New().Post(baseUrl + "/v1/object").End()
		if len(errs) > 0 {
			t.Fatalf("Error(s) getting /: %+v", errs)
		}
		if response.StatusCode/100 != 2 {
			t.Fatalf("Expected 2xx status-code but actual=%v; body=%v", response.StatusCode, body)
		}
		if expected, actual := `{"Success":true}`, body; actual != expected {
			t.Errorf("Expected /v1/object response body=%v but actual=%v", expected, actual)
		}
	}

	{
		response, body, errs := gorequest.New().Post(baseUrl + "/v1/objects").End()
		if len(errs) > 0 {
			t.Fatalf("Error(s) getting /: %+v", errs)
		}
		if response.StatusCode/100 != 2 {
			t.Fatalf("Expected 2xx status-code but actual=%v; body=%v", response.StatusCode, body)
		}
		if expected, actual := `{"meta":{"totalCount":4},"objects":["a","b","c","d"]}`, body; actual != expected {
			t.Errorf("Expected /v1/objects response body=%v but actual=%v", expected, actual)
		}
	}
}
