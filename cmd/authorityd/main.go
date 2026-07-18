package main

import (
	"encoding/base32"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/autotaker/codex-authority-broker/internal/ipc"
	"github.com/autotaker/codex-authority-broker/internal/lease"
	"github.com/autotaker/codex-authority-broker/internal/totp"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type config struct {
	Socket        string `json:"socket"`
	CodexUID      int    `json:"codex_uid"`
	SeedFile      string `json:"totp_seed_file"`
	ReplayKeyFile string `json:"replay_key_file"`
	ReplayFile    string `json:"replay_file"`
	GitHubKeyFile string `json:"github_key_file"`
}

func main() {
	p := flag.String("config", "/etc/codex-authority/config.json", "config")
	flag.Parse()
	if run(*p) != nil {
		fmt.Fprintln(os.Stderr, "authorityd: startup failed")
		os.Exit(1)
	}
}
func run(p string) error {
	b, e := protected(p, 4096)
	if e != nil {
		return e
	}
	defer zero(b)
	var c config
	if e = json.Unmarshal(b, &c); e != nil {
		return e
	}
	s, e := protected(c.SeedFile, 1024)
	if e != nil {
		return e
	}
	defer zero(s)
	seed, e := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(string(s))))
	if e != nil {
		return e
	}
	defer zero(seed)
	key, e := protected(c.ReplayKeyFile, 1024)
	if e != nil {
		return e
	}
	defer zero(key)
	app, e := protected(c.GitHubKeyFile, 65536)
	if e != nil {
		return e
	}
	zero(app)
	r, e := totp.NewFileReplay(c.ReplayFile, key)
	if e != nil {
		return e
	}
	v := totp.New(seed, time.Now, 1)
	defer v.Close()
	m, e := lease.New(lease.Config{Clock: lease.NewMonotonicClock(), Validator: v, Replay: r})
	if e != nil {
		return e
	}
	srv, e := ipc.Listen(c.Socket, 0660, handler(c.CodexUID, m))
	if e != nil {
		return e
	}
	defer srv.Close()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sig; srv.Close() }()
	e = srv.Serve()
	if strings.Contains(fmt.Sprint(e), "closed network connection") {
		return nil
	}
	return e
}
func handler(uid int, m *lease.Manager) ipc.Handler {
	return func(p ipc.Peer, q ipc.Message) ipc.Message {
		deny := ipc.Message{Type: q.Type, Status: ipc.StatusDenied}
		switch q.Type {
		case ipc.TypeCheck:
			if p.UID == 0 && m.Active() {
				return ipc.Message{Type: q.Type, Status: ipc.StatusOK}
			}
		case ipc.TypeConfirmReady:
			if p.UID == uid {
				s, e := m.ConfirmReady()
				if e == nil {
					return ipc.Message{Type: q.Type, Status: ipc.StatusOK, Payload: []byte(s.Challenge)}
				}
			}
		case ipc.TypeSubmitOTP:
			if p.UID == uid {
				x := strings.SplitN(string(q.Payload), "\n", 2)
				zero(q.Payload)
				if len(x) == 2 {
					otp := []byte(x[1])
					_, e := m.SubmitOTP(x[0], otp)
					zero(otp)
					if e == nil {
						return ipc.Message{Type: q.Type, Status: ipc.StatusOK}
					}
				}
			}
		case ipc.TypeAuthorizePush:
			if p.UID == uid && m.Active() {
				return ipc.Message{Type: q.Type, Status: ipc.StatusOK}
			}
		}
		return deny
	}
}
func protected(p string, max int64) ([]byte, error) {
	f, e := os.Open(p)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil {
		return nil, e
	}
	if st.Mode().Perm()&0077 != 0 || st.Size() > max {
		return nil, fmt.Errorf("unsafe secret file")
	}
	if x, ok := st.Sys().(*syscall.Stat_t); ok && os.Geteuid() == 0 && x.Uid != 0 {
		return nil, fmt.Errorf("unsafe secret owner")
	}
	b := make([]byte, st.Size())
	n, e := f.Read(b)
	if e != nil || int64(n) != st.Size() {
		return nil, fmt.Errorf("secret read failed")
	}
	return b, nil
}
func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
