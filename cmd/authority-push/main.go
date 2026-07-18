package main

import (
	"flag"
	"fmt"
	"github.com/autotaker/codex-authority-broker/internal/ipc"
	"os"
	"regexp"
)

var allowed = regexp.MustCompile(`^(main|task/TASK-[A-Za-z0-9][A-Za-z0-9._-]*)$`)

func main() {
	p := flag.String("socket", "/run/codex-authority/authority.sock", "socket")
	b := flag.String("branch", "", "branch")
	flag.Parse()
	if flag.NArg() != 0 || !allowed.MatchString(*b) {
		deny()
	}
	r, e := ipc.Call(*p, ipc.Message{Type: ipc.TypeAuthorizePush, Payload: []byte(*b)})
	if e != nil || r.Status != ipc.StatusOK {
		deny()
	}
	fmt.Println("push authority granted; transport unavailable")
}
func deny() { fmt.Fprintln(os.Stderr, "push denied"); os.Exit(1) }
