package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"github.com/autotaker/codex-authority-broker/internal/ipc"
	"os"
)

type req struct{ Method, Challenge, OTP string }
type rsp struct {
	OK        bool   `json:"ok"`
	Challenge string `json:"challenge,omitempty"`
	Error     string `json:"error,omitempty"`
}

func main() {
	p := flag.String("socket", "/run/codex-authority/authority.sock", "socket")
	flag.Parse()
	s := bufio.NewScanner(os.Stdin)
	s.Buffer(make([]byte, 1024), 4096)
	e := json.NewEncoder(os.Stdout)
	for s.Scan() {
		var q req
		if json.Unmarshal(s.Bytes(), &q) != nil {
			e.Encode(rsp{Error: "request denied"})
			continue
		}
		m := ipc.Message{}
		switch q.Method {
		case "confirm_ready":
			m.Type = ipc.TypeConfirmReady
		case "submit_otp":
			m.Type = ipc.TypeSubmitOTP
			m.Payload = []byte(q.Challenge + "\n" + q.OTP)
			q.OTP = ""
		default:
			e.Encode(rsp{Error: "request denied"})
			continue
		}
		r, x := ipc.Call(*p, m)
		zero(m.Payload)
		if x != nil || r.Status != ipc.StatusOK {
			e.Encode(rsp{Error: "request denied"})
			continue
		}
		o := rsp{OK: true}
		if q.Method == "confirm_ready" {
			o.Challenge = string(r.Payload)
		}
		e.Encode(o)
	}
}
func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
