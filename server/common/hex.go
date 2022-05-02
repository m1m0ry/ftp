package common

import (
	"encoding/hex"
	"hash"
	"io"
)

//哈希
type Hasher struct {
	io.Writer
	io.Reader
	hash.Hash
	Size uint64
}

func (h *Hasher) Write(p []byte) (n int, err error) {
	n, err = h.Writer.Write(p)
	h.Hash.Write(p)
	h.Size += uint64(n)
	return

}

func (h *Hasher) Read(p []byte) (n int, err error) {
	n, err = h.Reader.Read(p)
	h.Hash.Write(p[:n])
	return
}

func (h *Hasher) Sum() string {
	return hex.EncodeToString(h.Hash.Sum(nil))
}
