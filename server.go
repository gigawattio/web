package web

import (
	"crypto/tls"
	"fmt"
	golog "log"
	"net"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gigawattio/go-commons/pkg/errorlib"
	"github.com/jaytaylor/stoppableListener"
)

const (
	MaxStopChecks          = 10
	listenerStoppedMessage = "listener stopped"
)

type WebServerOptions struct {
	Addr           string        // TCP address to listen on, ":http" if empty.
	Handler        http.Handler  // handler to invoke, http.DefaultServeMux if nil.
	ReadTimeout    time.Duration // maximum duration before timing out read of the request.
	WriteTimeout   time.Duration // maximum duration before timing out write of the response.
	MaxHeaderBytes int           // maximum size of request headers, net/http.DefaultMaxHeaderBytes if 0.
	TLSConfig      *tls.Config   // optional TLS config, used by ListenAndServeTLS.
	ErrorLog       *golog.Logger
}

type WebServer struct {
	Options  WebServerOptions
	server   *http.Server
	listener *stoppableListener.StoppableListener
	lock     sync.RWMutex
}

type StaticHttpHandler struct {
	Content    []byte
	StatusCode int
	Headers    map[string]string
}

// NewWebServer creates a basic stoppable WebServer.
func NewWebServer(options WebServerOptions) *WebServer {
	ws := &WebServer{
		Options: options,
	}
	return ws
}

// NewStaticWebServer creates a WebServer which always serves the same content.
func NewStaticWebServer(options WebServerOptions, content []byte, statusCode int, headers map[string]string) *WebServer {
	options.Handler = StaticHandlerFunc(content, statusCode, headers)
	ws := &WebServer{
		Options: options,
	}
	return ws
}

// Start starts up the WebServer.
func (ws *WebServer) Start() error {
	ws.lock.Lock()
	defer ws.lock.Unlock()

	if ws.server != nil || ws.listener != nil {
		return errorlib.AlreadyRunningError
	}
	rawListener, err := net.Listen("tcp", ws.Options.Addr)
	if err != nil {
		return err
	}
	listener, err := stoppableListener.New(rawListener)
	if err != nil {
		return err
	}
	ws.listener = listener
	ws.server = &http.Server{
		Handler:        ws.Options.Handler,
		ReadTimeout:    ws.Options.ReadTimeout,
		WriteTimeout:   ws.Options.WriteTimeout,
		MaxHeaderBytes: ws.Options.MaxHeaderBytes,
		TLSConfig:      ws.Options.TLSConfig,
		ErrorLog:       ws.Options.ErrorLog,
	}
	go func() {
		if err := ws.server.Serve(ws.listener); err != nil && err != stoppableListener.StoppedError {
			log.Infof("web.WebServer: error on ws with Options=%+v: %s", ws.Options, err)
		}
		// log.Info("Server done!")
	}()
	return nil
}

// Stop terminates the WebServer.
func (ws *WebServer) Stop() error {
	ws.lock.Lock()
	defer ws.lock.Unlock()

	if ws.server == nil || ws.listener == nil {
		return errorlib.NotRunningError
	}
	if err := ws.listener.StopSafely(); err != nil {
		return err
	}
	ws.server = nil
	ws.listener = nil
	return nil
}

// Addr exposes the listener address.
func (ws *WebServer) Addr() net.Addr {
	ws.lock.RLock()
	defer ws.lock.RUnlock()

	if ws.listener == nil {
		return &net.IPAddr{}
	}
	addr := ws.listener.Addr()
	return addr
}

// BaseUrl provides a working URL base path to the web server instance.
func (ws *WebServer) BaseUrl() string {
	url := fmt.Sprintf("http://%s", ws.Addr())
	return url
}

// StaticHandlerFunc generates a handler which always serves up the same thing
// regardless of the request.
func StaticHandlerFunc(content []byte, statusCode int, headers map[string]string) http.HandlerFunc {
	log.Infof("Creating new static handler func with headers=%+v and content=%s", headers, string(content))
	handlerFn := func(w http.ResponseWriter, req *http.Request) {
		log.Infof("In the static handler where headers=%+v and content=%s", headers, string(content))
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(statusCode)
		w.Write(content)
	}
	return handlerFn
}
