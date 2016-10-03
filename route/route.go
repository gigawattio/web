package route

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/nbio/hitch"
)

// RouteMiddlewareBundle is the struct which represents a group of
// middleware + route entries.
type RouteMiddlewareBundle struct {
	Middlewares []func(http.Handler) http.Handler
	RouteData   []RouteDatum
}

// RouteDatum encompasses a single route entry.
type RouteDatum struct {
	Reciever    string // One of: "get", "post", "put", "patch", or "delete".  Or a combination of them separated by pipes, e.g.: "post|put"
	Path        string
	HandlerFunc func(w http.ResponseWriter, req *http.Request)
}

type HttpMethodReceiver func(path string, handler http.Handler, middleware ...func(http.Handler) http.Handler)

// Activate prepares a hitch for a single RouteMiddlewareBundle.
func (rmb *RouteMiddlewareBundle) Activate() *hitch.Hitch {
	h := hitch.New()
	h.Use(rmb.Middlewares...)
	for _, routeDatum := range rmb.RouteData {
		for _, method := range strings.Split(routeDatum.Reciever, "|") {
			receiverFunc := rmb.lookupReceiver(h, method)
			receiverFunc(routeDatum.Path, http.HandlerFunc(routeDatum.HandlerFunc))
			// log.Info("route: registered method=%s path=%s", method, routeDatum.Path)
		}
	}
	return h
}

// lookupReceiver takes an HTTP method string and returns the corresponding
// hitch.{method} function.
func (rmb *RouteMiddlewareBundle) lookupReceiver(h *hitch.Hitch, method string) HttpMethodReceiver {
	switch strings.ToLower(method) {
	case "get":
		return h.Get
	case "post":
		return h.Post
	case "put":
		return h.Put
	case "patch":
		return h.Patch
	case "delete":
		return h.Delete
	default:
		panic(fmt.Sprintf("unable to resolve route for http method: %s", method))
	}
}

// Activate hitches one or more RouteMiddlewareBundle structs together.
func Activate(rmbs []RouteMiddlewareBundle) *hitch.Hitch {
	var head *hitch.Hitch
	var tail *hitch.Hitch // Used to auto-link hitches together.
	for _, rmb := range rmbs {
		h := rmb.Activate()
		if head == nil {
			head = h
		}
		if tail != nil {
			tail.Next(h.Handler())
		}
		tail = h
	}
	return head
}
