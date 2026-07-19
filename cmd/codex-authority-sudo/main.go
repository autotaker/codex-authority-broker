package main

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/autotaker/codex-authority-broker/internal/ipc"
)

const (
	defaultSocketPath = "/run/codex-authority.sock"
	deniedLine        = "request denied\n"
	authorityTimeout  = 2 * time.Second
)

// callFunc is kept injectable for deterministic unit and isolated-fixture
// tests. Production uses ipc.Client.Call directly.
type callFunc func(context.Context, ipc.Request) (ipc.Response, error)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, nil))
}

// run is the pam_exec-compatible entrypoint. It deliberately accepts no
// command-line or stdin authority material: the only decision input is one
// payload-free authorize request to the fixed broker socket.
func run(args []string, _ io.Reader, stdout, stderr io.Writer, call callFunc) int {
	// pam_exec may append configured arguments. They are deliberately ignored:
	// no argument is an action, authority, or privileged-command input, and
	// every process invocation still performs exactly one live check.
	_ = args

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
