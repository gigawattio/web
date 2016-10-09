package web

import (
	"crypto/tls"
	"errors"
	"fmt"
	golog "log"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gigawattio/go-commons/pkg/errorlib"
	"github.com/jaytaylor/stoppableListener"
)

const (
	listenerStoppedMessage = "listener stopped"
	MaxStopChecks          = 10
)

var (
	ListenerAlreadyInUseError = errors.New("listener already in use")
	NilServerError            = errors.New("server not in use")
	NotStoppedError           = fmt.Errorf("server failed to stop, port is still open after %v checks", MaxStopChecks)
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
	lock     sync.Mutex
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

	if ws.listener != nil {
		return ListenerAlreadyInUseError
	}
	if ws.server != nil {
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
		if err := ws.server.Serve(ws.listener); err != nil && err.Error() != listenerStoppedMessage {
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

	if ws.server == nil {
		return errorlib.NotRunningError
	}
	if ws.listener == nil {
		return ListenerAlreadyInUseError
	}
	ws.listener.Stop()
	time.Sleep(50 * time.Millisecond)
	if err := ws.waitUntilStopped(); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	ws.listener = nil
	ws.server = nil
	return nil
}

// Addr exposes the listener address.
func (ws *WebServer) Addr() net.Addr {
	ws.lock.Lock()
	defer ws.lock.Unlock()

	if ws.listener == nil {
		return &net.IPAddr{}
	}
	addr := ws.listener.Addr()
	return addr
}

// waitUntilStopped uses netcat (nc) to determine if the listening port is
// still accepting connections.  Returns nil when connections are no longer
// being accepted, or returns NotStoppedError if MaxStopChecks are exceeded.
func (ws *WebServer) waitUntilStopped() error {
	args := append([]string{"-v", "-w", "1"}, strings.Split(ws.Options.Addr, ":")...)
	for i := 0; i < MaxStopChecks; i++ {
		/*out*/ _, err := exec.Command("nc", args...).CombinedOutput()
		if err != nil { // If `nc` exits with non-zero status code then that means the port is closed.
			return nil
		}
		/*log.Printf("waitUntilStopped nc output=%s\n", string(out))*/
	}
	return NotStoppedError
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
