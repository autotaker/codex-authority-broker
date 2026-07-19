//go:build linux

package ipc

import (
	"context"
	"net"
	"path/filepath"
	"time"
)

type Client struct {
	Path string
}

func (c Client) Call(ctx context.Context, request Request) (Response, error) {
	if !filepath.IsAbs(c.Path) || filepath.Clean(c.Path) != c.Path {
		return Response{}, ErrTransport
	}
	connection, err := (&net.Dialer{}).DialContext(ctx, "unix", c.Path)
	if err != nil {
		return Response{}, ErrTransport
	}
	defer connection.Close()
	deadline := time.Now().Add(ioDeadline)
	if contextDeadline, ok := ctx.Deadline(); ok && contextDeadline.Before(deadline) {
		deadline = contextDeadline
	}
	if err := connection.SetDeadline(deadline); err != nil {
		return Response{}, ErrTransport
	}
	if err := writeRequest(connection, request); err != nil {
		return Response{}, err
	}
	response, err := readResponse(connection)
	if err != nil {
		return Response{}, err
	}
	return response, nil
}
