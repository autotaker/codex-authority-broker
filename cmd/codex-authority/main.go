package main

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/autotaker/codex-authority-broker/internal/ipc"
)

const (
	defaultSocketPath = "/run/codex-authority.sock"
	deniedLine        = "request denied\n"
)

type callFunc func(context.Context, ipc.Request) (ipc.Response, error)

func main() {
	client := ipc.Client{Path: socketPath(os.Args[1:])}
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, client.Call))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, call callFunc) int {
	path := defaultSocketPath
	if len(args) >= 2 && args[0] == "--socket" {
		path = args[1]
		args = args[2:]
	}
	if len(args) != 1 || (args[0] != ipc.OperationReady && args[0] != ipc.OperationOTP) {
		_, _ = io.WriteString(stderr, deniedLine)
		return 1
	}
	request := ipc.Request{Version: ipc.ProtocolVersion, Operation: args[0]}
	if args[0] == ipc.OperationOTP {
		code, ok := readOTP(stdin)
		if !ok {
			_, _ = io.WriteString(stderr, deniedLine)
			return 1
		}
		payload, err := json.Marshal(struct {
			Code string `json:"code"`
		}{Code: code})
		if err != nil {
			_, _ = io.WriteString(stderr, deniedLine)
			return 1
		}
		request.Payload = payload
	}
	client := ipc.Client{Path: path}
	response, err := callWithClient(call, client, request)
	if err != nil || !response.OK {
		_, _ = io.WriteString(stderr, deniedLine)
		return 1
	}
	_, _ = io.WriteString(stdout, args[0]+" accepted\n")
	return 0
}

func callWithClient(call callFunc, client ipc.Client, request ipc.Request) (ipc.Response, error) {
	if call != nil {
		return call(context.Background(), request)
	}
	return client.Call(context.Background(), request)
}

func socketPath(args []string) string {
	if len(args) >= 2 && args[0] == "--socket" {
		return args[1]
	}
	return defaultSocketPath
}

func readOTP(stdin io.Reader) (string, bool) {
	input, err := io.ReadAll(io.LimitReader(stdin, 9))
	if err != nil || len(input) < 2 || input[len(input)-1] != '\n' {
		return "", false
	}
	input = input[:len(input)-1]
	if len(input) > 0 && input[len(input)-1] == '\r' {
		input = input[:len(input)-1]
	}
	if len(input) != 6 {
		return "", false
	}
	for _, digit := range input {
		if digit < '0' || digit > '9' {
			return "", false
		}
	}
	return string(input), true
}
