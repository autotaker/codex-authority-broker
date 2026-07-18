package totp

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type FileReplay struct {
	mu   sync.Mutex
	path string
	key  []byte
}

func NewFileReplay(p string, k []byte) (*FileReplay, error) {
	if len(k) < 32 {
		return nil, errors.New("replay key too short")
	}
	return &FileReplay{path: p, key: append([]byte(nil), k...)}, nil
}
func (r *FileReplay) fp(s uint64) string {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], s)
	h := hmac.New(sha256.New, r.key)
	h.Write(b[:])
	return hex.EncodeToString(h.Sum(nil))
}
func (r *FileReplay) has(s uint64) (bool, error) {
	f, e := os.Open(r.path)
	if os.IsNotExist(e) {
		return false, nil
	}
	if e != nil {
		return false, e
	}
	defer f.Close()
	n := r.fp(s)
	q := bufio.NewScanner(f)
	for q.Scan() {
		if hmac.Equal([]byte(q.Text()), []byte(n)) {
			return true, nil
		}
	}
	return false, q.Err()
}
func (r *FileReplay) Consumed(s uint64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.has(s)
}
func (r *FileReplay) Consume(s uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	yes, e := r.has(s)
	if e != nil || yes {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(r.path), 0700); e != nil {
		return e
	}
	f, e := os.OpenFile(r.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	if _, e = f.WriteString(r.fp(s) + "\n"); e != nil {
		return e
	}
	return f.Sync()
}
