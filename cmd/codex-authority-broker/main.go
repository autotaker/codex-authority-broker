//go:build linux

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/autotaker/codex-authority-broker/internal/backend"
	"github.com/autotaker/codex-authority-broker/internal/ipc"
)

const (
	seedPath          = "/etc/codex-authority/seed.json"
	socketPath        = "/run/codex-authority.sock"
	maxSeedBytes      = 1024
	maxSecretBytes    = 128
	maxSeedComponents = 16
	maxSeedPathBytes  = 256
	maxComponentBytes = 255
	directoryFlags    = syscall.O_RDONLY | syscall.O_DIRECTORY | syscall.O_NOFOLLOW | syscall.O_CLOEXEC
	finalFileFlags    = syscall.O_RDONLY | syscall.O_NOFOLLOW | syscall.O_CLOEXEC
)

var errSeed = errors.New("seed unavailable")

type descriptorStat struct {
	mode uint32
	uid  uint32
	size int64
}

type seedDescriptors interface {
	openRoot(int) (int, error)
	openAt(int, string, int) (int, error)
	stat(int) (descriptorStat, error)
	close(int) error
	file(int) (io.ReadCloser, error)
}

type systemDescriptors struct{}

func (systemDescriptors) openRoot(flags int) (int, error) { return syscall.Open("/", flags, 0) }

func (systemDescriptors) openAt(dirfd int, name string, flags int) (int, error) {
	return syscall.Openat(dirfd, name, flags, 0)
}

func (systemDescriptors) stat(fd int) (descriptorStat, error) {
	var value syscall.Stat_t
	if err := syscall.Fstat(fd, &value); err != nil {
		return descriptorStat{}, err
	}
	return descriptorStat{mode: value.Mode, uid: value.Uid, size: value.Size}, nil
}

func (systemDescriptors) close(fd int) error { return syscall.Close(fd) }

func (systemDescriptors) file(fd int) (io.ReadCloser, error) {
	file := os.NewFile(uintptr(fd), "")
	if file == nil {
		return nil, errSeed
	}
	return file, nil
}

type brokerRuntime interface {
	ipc.Backend
	Close()
}

type brokerServer interface {
	Serve(context.Context) error
	Close() error
}

type runtimeFactory func([]byte) (brokerRuntime, error)
type listenFactory func(ipc.Config, ipc.Backend) (brokerServer, error)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	status := run(ctx, systemDescriptors{}, realRuntime, realListen)
	stop()
	os.Exit(status)
}

func realRuntime(secret []byte) (brokerRuntime, error) { return backend.New(secret) }

func realListen(config ipc.Config, runtime ipc.Backend) (brokerServer, error) {
	return ipc.Listen(config, runtime)
}

func run(ctx context.Context, descriptors seedDescriptors, makeRuntime runtimeFactory, listen listenFactory) (status int) {
	if ctx == nil || ctx.Err() != nil || makeRuntime == nil || listen == nil {
		return 1
	}
	secret, uid, err := loadSeed(descriptors)
	if err != nil {
		return 1
	}
	runtime, err := makeRuntime(secret)
	wipe(secret)
	if err != nil || runtime == nil {
		return 1
	}
	server, err := listen(ipc.Config{Path: socketPath, AllowedUID: uid}, runtime)
	if err != nil || server == nil {
		if server != nil {
			_ = server.Close()
		}
		runtime.Close()
		return 1
	}
	serveErr, cancelled := ctx.Err(), false
	defer func() {
		runtime.Close()
		if closeErr := server.Close(); serveErr != nil || !cancelled || closeErr != nil {
			status = 1
		}
	}()
	serveErr = server.Serve(ctx)
	cancelled = ctx.Err() != nil
	return 0
}

func loadSeed(descriptors seedDescriptors) ([]byte, uint32, error) {
	fd, size, err := openSeed(descriptors)
	if err != nil {
		return nil, 0, errSeed
	}
	reader, err := descriptors.file(fd)
	if err != nil || reader == nil {
		_ = descriptors.close(fd)
		return nil, 0, errSeed
	}
	data, readErr := readSeed(reader, size)
	if closeErr := reader.Close(); readErr != nil || closeErr != nil {
		wipe(data)
		return nil, 0, errSeed
	}
	secret, uid, parseErr := parseSeed(data)
	wipe(data)
	if parseErr != nil {
		wipe(secret)
		return nil, 0, errSeed
	}
	return secret, uid, nil
}

func openSeed(descriptors seedDescriptors) (int, int64, error) {
	components, ok := seedComponents(seedPath)
	if !ok || descriptors == nil {
		return -1, 0, errSeed
	}
	current, err := descriptors.openRoot(directoryFlags)
	if err != nil {
		return -1, 0, errSeed
	}
	root, err := descriptors.stat(current)
	if err != nil || !isDirectory(root.mode) {
		_ = descriptors.close(current)
		return -1, 0, errSeed
	}
	for index, component := range components {
		flags := directoryFlags
		if index == len(components)-1 {
			flags = finalFileFlags
		}
		next, openErr := descriptors.openAt(current, component, flags)
		if openErr != nil {
			_ = descriptors.close(current)
			return -1, 0, errSeed
		}
		metadata, statErr := descriptors.stat(next)
		if statErr != nil {
			_ = descriptors.close(next)
			_ = descriptors.close(current)
			return -1, 0, errSeed
		}
		if index == len(components)-1 {
			if descriptors.close(current) != nil || !validSeedFile(metadata) {
				_ = descriptors.close(next)
				return -1, 0, errSeed
			}
			return next, metadata.size, nil
		}
		if descriptors.close(current) != nil || !isDirectory(metadata.mode) {
			_ = descriptors.close(next)
			return -1, 0, errSeed
		}
		current = next
	}
	return -1, 0, errSeed
}

func seedComponents(path string) ([]string, bool) {
	if len(path) < 2 || len(path) > maxSeedPathBytes || path[0] != '/' || strings.Contains(path[1:], "//") {
		return nil, false
	}
	components := strings.Split(path[1:], "/")
	if len(components) > maxSeedComponents {
		return nil, false
	}
	for _, component := range components {
		if component == "" || component == "." || component == ".." || len(component) > maxComponentBytes || strings.ContainsAny(component, "/\\") {
			return nil, false
		}
	}
	return components, true
}

func isDirectory(mode uint32) bool { return mode&syscall.S_IFMT == syscall.S_IFDIR }

func validSeedFile(metadata descriptorStat) bool {
	return metadata.mode&syscall.S_IFMT == syscall.S_IFREG && metadata.uid == 0 && metadata.mode&07777 == 0600 && metadata.size > 0 && metadata.size <= maxSeedBytes
}

func readSeed(reader io.Reader, size int64) ([]byte, error) {
	if size < 1 || size > maxSeedBytes {
		return nil, errSeed
	}
	data, err := io.ReadAll(io.LimitReader(reader, maxSeedBytes+1))
	if err != nil || int64(len(data)) != size || len(data) > maxSeedBytes {
		return data, errSeed
	}
	return data, nil
}

func parseSeed(document []byte) (secret []byte, uid uint32, err error) {
	defer func() {
		if err != nil {
			wipe(secret)
			secret = nil
		}
	}()
	if len(document) == 0 || len(document) > maxSeedBytes {
		return nil, 0, errSeed
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	start, tokenErr := decoder.Token()
	if tokenErr != nil || start != json.Delim('{') {
		return nil, 0, errSeed
	}
	var haveSecret, haveUID bool
	for decoder.More() {
		key, keyOK := decoder.Token()
		name, nameOK := key.(string)
		if keyOK != nil || !nameOK {
			return nil, 0, errSeed
		}
		if name == "totp_secret_b64" {
			if haveSecret {
				return nil, 0, errSeed
			}
			haveSecret = true
			value, valueErr := decoder.Token()
			encoded, valueOK := value.(string)
			if valueErr != nil || !valueOK || len(encoded) == 0 || len(encoded) > base64.StdEncoding.EncodedLen(maxSecretBytes) {
				return nil, 0, errSeed
			}
			decoded, decodeErr := base64.StdEncoding.DecodeString(encoded)
			if decodeErr != nil || len(decoded) == 0 || len(decoded) > maxSecretBytes || base64.StdEncoding.EncodeToString(decoded) != encoded {
				return nil, 0, errSeed
			}
			secret = decoded
			continue
		}
		if name == "allowed_uid" {
			if haveUID {
				return nil, 0, errSeed
			}
			haveUID = true
			value, valueErr := decoder.Token()
			number, valueOK := value.(json.Number)
			if valueErr != nil || !valueOK {
				return nil, 0, errSeed
			}
			parsed, parseErr := strconv.ParseUint(number.String(), 10, 32)
			if parseErr != nil || parsed == 0 {
				return nil, 0, errSeed
			}
			uid = uint32(parsed)
			continue
		}
		return nil, 0, errSeed
	}
	end, endErr := decoder.Token()
	if endErr != nil || end != json.Delim('}') || !haveSecret || !haveUID {
		return nil, 0, errSeed
	}
	if _, trailingErr := decoder.Token(); trailingErr != io.EOF {
		return nil, 0, errSeed
	}
	return secret, uid, nil
}

func wipe(data []byte) { clear(data) }
