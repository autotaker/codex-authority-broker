package ipc

import (
	"errors"
	"net"
	"os"
	"syscall"
	"time"
)

type Peer struct{ PID, UID, GID int }
type Handler func(Peer, Message) Message
type Server struct {
	path string
	ln   *net.UnixListener
	h    Handler
}

func Listen(p string, mode os.FileMode, h Handler) (*Server, error) {
	if h == nil {
		return nil, errors.New("nil handler")
	}
	os.Remove(p)
	l, e := net.ListenUnix("unixpacket", &net.UnixAddr{Name: p, Net: "unixpacket"})
	if e != nil {
		return nil, e
	}
	if e = os.Chmod(p, mode); e != nil {
		l.Close()
		return nil, e
	}
	return &Server{p, l, h}, nil
}
func (s *Server) Serve() error {
	for {
		c, e := s.ln.AcceptUnix()
		if e != nil {
			return e
		}
		go s.one(c)
	}
}
func (s *Server) one(c *net.UnixConn) {
	defer c.Close()
	raw, e := c.SyscallConn()
	if e != nil {
		return
	}
	var u *syscall.Ucred
	raw.Control(func(fd uintptr) { u, e = syscall.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED) })
	if e != nil {
		return
	}
	b := make([]byte, HeaderSize+MaxPayload+1)
	n, _, _, _, e := c.ReadMsgUnix(b, nil)
	if e != nil || n > HeaderSize+MaxPayload {
		return
	}
	m, e := Decode(b[:n])
	if e != nil {
		return
	}
	out, e := Encode(s.h(Peer{int(u.Pid), int(u.Uid), int(u.Gid)}, m))
	if e == nil {
		c.Write(out)
	}
}
func (s *Server) Close() error { os.Remove(s.path); return s.ln.Close() }
func Call(p string, m Message) (Message, error) {
	c, e := net.DialUnix("unixpacket", nil, &net.UnixAddr{Name: p, Net: "unixpacket"})
	if e != nil {
		return Message{}, e
	}
	defer c.Close()
	c.SetDeadline(time.Now().Add(2 * time.Second))
	b, e := Encode(m)
	if e != nil {
		return Message{}, e
	}
	if _, e = c.Write(b); e != nil {
		return Message{}, e
	}
	r := make([]byte, HeaderSize+MaxPayload+1)
	n, e := c.Read(r)
	if e != nil {
		return Message{}, e
	}
	return Decode(r[:n])
}
