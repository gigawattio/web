package web

import (
	"net"
	"net/http"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/gigawattio/go-commons/pkg/errorlib"
	"github.com/jaytaylor/stoppableListener"
)

// FsServer is a filesystem server.
type FsServer struct {
	bind     string
	dir      http.Dir
	server   *http.Server
	listener *stoppableListener.StoppableListener
	lock     sync.Mutex
}

func NewFsServer(bind string, dir http.Dir) *FsServer {
	fsServer := &FsServer{
		bind: bind,
		dir:  dir,
		server: &http.Server{
			Handler: http.FileServer(dir),
		},
	}
	return fsServer
}

func (fsServer *FsServer) Start() error {
	fsServer.lock.Lock()
	defer fsServer.lock.Unlock()

	if fsServer.listener != nil {
		return errorlib.AlreadyRunningError
	}

	rawListener, err := net.Listen("tcp", fsServer.bind)
	if err != nil {
		return err
	}
	sl, err := stoppableListener.New(rawListener)
	if err != nil {
		return err
	}
	fsServer.listener = sl
	go func() {
		if err := fsServer.server.Serve(sl); err != nil && err.Error() != listenerStoppedMessage {
			log.Errorf("unexpected error from FsServer.server.Serve(sl): %s", err)
		}
	}()
	return nil
}

func (fsServer *FsServer) Stop() error {
	fsServer.lock.Lock()
	defer fsServer.lock.Unlock()

	if fsServer.listener == nil {
		return errorlib.NotRunningError
	}

	if err := fsServer.listener.Stop(); err != nil {
		return err
	}
	fsServer.listener = nil
	return nil
}

func (fsServer *FsServer) Addr() net.Addr {
	fsServer.lock.Lock()
	defer fsServer.lock.Unlock()

	if fsServer.listener == nil {
		return &net.IPAddr{}
	}

	addr := fsServer.listener.Addr()
	return addr
}

func (fsServer *FsServer) Dir() http.Dir {
	return fsServer.dir
}
