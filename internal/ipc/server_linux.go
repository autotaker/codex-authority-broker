//go:build linux

package ipc

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	maxHandlers = 16
	ioDeadline  = 2 * time.Second
)

var (
	ErrServer     = errors.New("ipc: server unavailable")
	ErrUnsafePath = errors.New("ipc: unsafe socket path")
	ErrLifecycle  = errors.New("ipc: socket lifecycle failure")
)

type Backend interface {
	Handle(context.Context, Request) (Response, error)
}

type Config struct {
	Path       string
	AllowedUID uint32
}

type credentialReader func(*net.UnixConn) (uint32, error)

type socketIdentity struct {
	device uint64
	inode  uint64
}

type Server struct {
	listener   *net.UnixListener
	path       string
	identity   socketIdentity
	allowedUID uint32
	backend    Backend
	credential credentialReader

	mu          sync.Mutex
	closing     bool
	serving     bool
	connections map[*net.UnixConn]struct{}
	handlers    chan struct{}
	wait        sync.WaitGroup
	closeOnce   sync.Once
	closeDone   chan struct{}
	closeErr    error
	context     context.Context
	cancel      context.CancelFunc
}

func Listen(config Config, backend Backend) (*Server, error) {
	return listen(config, backend, kernelPeerUID)
}

func listen(config Config, backend Backend, credential credentialReader) (*Server, error) {
	if backend == nil || credential == nil {
		return nil, ErrServer
	}
	if err := validateSocketPath(config.Path); err != nil {
		return nil, err
	}
	if _, err := os.Lstat(config.Path); err == nil || !os.IsNotExist(err) {
		return nil, ErrUnsafePath
	}
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: config.Path, Net: "unix"})
	if err != nil {
		return nil, ErrServer
	}
	listener.SetUnlinkOnClose(false)
	identity, err := identifySocket(config.Path)
	if err != nil {
		_ = listener.Close()
		return nil, ErrLifecycle
	}
	fail := func() (*Server, error) {
		_ = listener.Close()
		_ = removeOwnedSocket(config.Path, identity)
		return nil, ErrLifecycle
	}
	if err := os.Chmod(config.Path, 0o600); err != nil {
		return fail()
	}
	serverContext, cancel := context.WithCancel(context.Background())
	return &Server{
		listener: listener, path: config.Path, identity: identity,
		allowedUID: config.AllowedUID, backend: backend, credential: credential,
		connections: make(map[*net.UnixConn]struct{}), handlers: make(chan struct{}, maxHandlers),
		closeDone: make(chan struct{}), context: serverContext, cancel: cancel,
	}, nil
}

func (s *Server) Serve(ctx context.Context) error {
	s.mu.Lock()
	if s.serving || s.closing {
		s.mu.Unlock()
		return ErrServer
	}
	s.serving = true
	s.mu.Unlock()

	stopWatcher := make(chan struct{})
	defer close(stopWatcher)
	go func() {
		select {
		case <-ctx.Done():
			_ = s.Close()
		case <-stopWatcher:
		}
	}()

	for {
		connection, err := s.listener.AcceptUnix()
		if err != nil {
			s.mu.Lock()
			closing := s.closing
			s.mu.Unlock()
			if closing || ctx.Err() != nil {
				return nil
			}
			return ErrServer
		}
		select {
		case s.handlers <- struct{}{}:
		default:
			_ = connection.Close()
			continue
		}
		s.mu.Lock()
		if s.closing {
			s.mu.Unlock()
			<-s.handlers
			_ = connection.Close()
			continue
		}
		s.connections[connection] = struct{}{}
		s.wait.Add(1)
		s.mu.Unlock()
		go s.handle(connection)
	}
}

func (s *Server) Close() error {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.closing = true
		s.cancel()
		listenerErr := s.listener.Close()
		for connection := range s.connections {
			_ = connection.Close()
		}
		s.mu.Unlock()
		s.wait.Wait()
		if listenerErr != nil && !errors.Is(listenerErr, net.ErrClosed) {
			s.closeErr = ErrLifecycle
		}
		if err := s.cleanup(); err != nil {
			s.closeErr = err
		}
		close(s.closeDone)
	})
	<-s.closeDone
	return s.closeErr
}

func (s *Server) handle(connection *net.UnixConn) {
	defer func() {
		_ = connection.Close()
		s.mu.Lock()
		delete(s.connections, connection)
		s.mu.Unlock()
		<-s.handlers
		s.wait.Done()
	}()
	_ = connection.SetDeadline(time.Now().Add(ioDeadline))
	uid, err := s.credential(connection)
	if err != nil || uid != s.allowedUID {
		return
	}
	request, err := readRequest(connection)
	if err != nil {
		s.writeFailure(connection)
		return
	}
	s.mu.Lock()
	if s.closing {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	response, backendErr := s.backend.Handle(s.context, request)
	if backendErr != nil {
		s.writeFailure(connection)
		return
	}
	_ = connection.SetWriteDeadline(time.Now().Add(ioDeadline))
	if err := writeResponse(connection, response); err != nil {
		return
	}
}

func (s *Server) writeFailure(connection *net.UnixConn) {
	_ = connection.SetWriteDeadline(time.Now().Add(ioDeadline))
	_ = writeResponse(connection, Response{OK: false})
}

func (s *Server) cleanup() error {
	return removeOwnedSocket(s.path, s.identity)
}

func removeOwnedSocket(path string, expected socketIdentity) error {
	identity, err := identifySocket(path)
	if err != nil || identity != expected {
		return ErrLifecycle
	}
	if err := os.Remove(path); err != nil {
		return ErrLifecycle
	}
	return nil
}

func validateSocketPath(path string) error {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || path == string(filepath.Separator) {
		return ErrUnsafePath
	}
	parent := filepath.Dir(path)
	current := string(filepath.Separator)
	for _, component := range strings.Split(strings.TrimPrefix(parent, current), current) {
		if component == "" {
			continue
		}
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return ErrUnsafePath
		}
	}
	info, err := os.Lstat(parent)
	if err != nil || !info.IsDir() || info.Mode().Perm()&0o022 != 0 {
		return ErrUnsafePath
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != uint32(os.Geteuid()) {
		return ErrUnsafePath
	}
	return nil
}

func identifySocket(path string) (socketIdentity, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return socketIdentity{}, err
	}
	if info.Mode()&os.ModeSocket == 0 {
		return socketIdentity{}, ErrLifecycle
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return socketIdentity{}, ErrLifecycle
	}
	return socketIdentity{device: uint64(stat.Dev), inode: stat.Ino}, nil
}

func kernelPeerUID(connection *net.UnixConn) (uint32, error) {
	raw, err := connection.SyscallConn()
	if err != nil {
		return 0, ErrServer
	}
	var uid uint32
	var credentialErr error
	if err := raw.Control(func(fd uintptr) {
		credential, err := syscall.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
		if err != nil {
			credentialErr = ErrServer
			return
		}
		uid = credential.Uid
	}); err != nil || credentialErr != nil {
		return 0, ErrServer
	}
	return uid, nil
}
