package main

import (
	"context"
	"io"
	"os"
	"syscall"
	"time"

	"github.com/autotaker/codex-authority-broker/internal/ipc"
)

const (
	defaultSocketPath = "/run/codex-authority.sock"
	defaultSocketDir  = "/run"
	deniedLine        = "request denied\n"
	authorityTimeout  = 2 * time.Second
)

// callFunc is kept injectable for deterministic unit and isolated-fixture
// tests. Production uses ipc.Client.Call directly.
type callFunc func(context.Context, ipc.Request) (ipc.Response, error)

// identityHooks keeps the privileged boundary deterministic in unit tests.
// Production always uses productionIdentityHooks; no caller-controlled value
// can replace any of these operations.
type identityHooks struct {
	lstat     func(string) (os.FileInfo, error)
	setgroups func([]int) error
	getgroups func() ([]int, error)
	setgid    func(int) error
	getgid    func() int
	getegid   func() int
	setuid    func(int) error
	getuid    func() int
	geteuid   func() int
}

var productionIdentityHooks = identityHooks{
	lstat:     os.Lstat,
	setgroups: syscall.Setgroups,
	getgroups: syscall.Getgroups,
	setgid:    syscall.Setgid,
	getgid:    syscall.Getgid,
	getegid:   syscall.Getegid,
	setuid:    syscall.Setuid,
	getuid:    syscall.Getuid,
	geteuid:   syscall.Geteuid,
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, nil))
}

// run is the pam_exec-compatible entrypoint. It deliberately accepts no
// command-line or stdin authority material: the only decision input is one
// payload-free authorize request to the fixed broker socket.
func run(args []string, _ io.Reader, stdout, stderr io.Writer, call callFunc) int {
	return runWithHooks(args, nil, stdout, stderr, call, productionIdentityHooks)
}

func runWithHooks(args []string, _ io.Reader, stdout, stderr io.Writer, call callFunc, hooks identityHooks) int {
	// pam_exec may append configured arguments. They are deliberately ignored:
	// no argument is an action, authority, or privileged-command input, and
	// every process invocation still performs exactly one live check.
	_ = args
	hooks = normalizeIdentityHooks(hooks)
	id, ok := fixedSocketIdentity(hooks.lstat)
	if !ok || !dropIdentity(id, hooks) {
		writeDenied(stderr)
		return 1
	}

	request := ipc.Request{
		Version:   ipc.ProtocolVersion,
		Operation: ipc.OperationAuthorize,
	}
	ctx, cancel := context.WithTimeout(context.Background(), authorityTimeout)
	defer cancel()
	response, err := callWithContext(ctx, call, ipc.Client{Path: defaultSocketPath}, request)
	if err != nil || response.Version != ipc.ProtocolVersion || !response.OK || len(response.Payload) != 0 {
		writeDenied(stderr)
		return 1
	}

	// pam_exec consumes the exit status. Keep both output streams empty on an
	// allow, so no lease, identity, socket, or backend detail can be disclosed.
	_ = stdout
	return 0
}

func normalizeIdentityHooks(hooks identityHooks) identityHooks {
	if hooks.lstat == nil {
		hooks.lstat = productionIdentityHooks.lstat
	}
	if hooks.setgroups == nil {
		hooks.setgroups = productionIdentityHooks.setgroups
	}
	if hooks.getgroups == nil {
		hooks.getgroups = productionIdentityHooks.getgroups
	}
	if hooks.setgid == nil {
		hooks.setgid = productionIdentityHooks.setgid
	}
	if hooks.getgid == nil {
		hooks.getgid = productionIdentityHooks.getgid
	}
	if hooks.getegid == nil {
		hooks.getegid = productionIdentityHooks.getegid
	}
	if hooks.setuid == nil {
		hooks.setuid = productionIdentityHooks.setuid
	}
	if hooks.getuid == nil {
		hooks.getuid = productionIdentityHooks.getuid
	}
	if hooks.geteuid == nil {
		hooks.geteuid = productionIdentityHooks.geteuid
	}
	return hooks
}

type fixedSocketStat struct {
	dev  uint64
	ino  uint64
	mode uint32
	uid  uint32
	gid  uint32
}

func fixedSocketIdentity(lstat func(string) (os.FileInfo, error)) (uint32, bool) {
	if lstat == nil {
		return 0, false
	}
	parent, err := lstat(defaultSocketDir)
	if err != nil {
		return 0, false
	}
	parentStat, ok := statFromFileInfo(parent)
	if !ok || parentStat.mode&syscall.S_IFMT != syscall.S_IFDIR || parentStat.uid != 0 || parentStat.mode&022 != 0 {
		return 0, false
	}
	first, err := lstat(defaultSocketPath)
	if err != nil {
		return 0, false
	}
	firstStat, ok := statFromFileInfo(first)
	if !ok || !validSocketStat(firstStat) {
		return 0, false
	}
	second, err := lstat(defaultSocketPath)
	if err != nil {
		return 0, false
	}
	secondStat, ok := statFromFileInfo(second)
	if !ok || !validSocketStat(secondStat) || firstStat != secondStat {
		return 0, false
	}
	return firstStat.uid, true
}

func statFromFileInfo(info os.FileInfo) (fixedSocketStat, bool) {
	if info == nil {
		return fixedSocketStat{}, false
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat == nil {
		return fixedSocketStat{}, false
	}
	return fixedSocketStat{
		dev:  uint64(stat.Dev),
		ino:  uint64(stat.Ino),
		mode: uint32(stat.Mode),
		uid:  uint32(stat.Uid),
		gid:  uint32(stat.Gid),
	}, true
}

func validSocketStat(stat fixedSocketStat) bool {
	return stat.mode&syscall.S_IFMT == syscall.S_IFSOCK && stat.uid != 0 && stat.gid != 0 && stat.uid == stat.gid
}

func dropIdentity(id uint32, hooks identityHooks) bool {
	if id == 0 {
		return false
	}
	if err := hooks.setgroups([]int{}); err != nil {
		return false
	}
	groups, err := hooks.getgroups()
	if err != nil || len(groups) != 0 {
		return false
	}
	if err := hooks.setgid(int(id)); err != nil {
		return false
	}
	if hooks.getgid() != int(id) || hooks.getegid() != int(id) {
		return false
	}
	if err := hooks.setuid(int(id)); err != nil {
		return false
	}
	return hooks.getuid() == int(id) && hooks.geteuid() == int(id)
}

func callWithContext(ctx context.Context, call callFunc, client ipc.Client, request ipc.Request) (ipc.Response, error) {
	if call != nil {
		return call(ctx, request)
	}
	return client.Call(ctx, request)
}

// callWithClient preserves the small seam used by sibling CLI tests while
// retaining one-call/no-retry semantics.
func callWithClient(call callFunc, client ipc.Client, request ipc.Request) (ipc.Response, error) {
	return callWithContext(context.Background(), call, client, request)
}

func writeDenied(stderr io.Writer) {
	if stderr != nil {
		_, _ = io.WriteString(stderr, deniedLine)
	}
}
