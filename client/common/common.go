package common

import (
	"encoding/hex"
	"fmt"
	"hash"
	"io"
)

//服务器地址
var BaseUrl string

// 文件元数据
type FileInfo struct {
	Filename string // 文件名
	Filesize int64  // 文件大小
	Filetype string // 文件类型（分为普通文件和临时文件）
}

// 文件列表
type ListFileInfos struct {
	Files []FileInfo
}

//下载进度
type Reader struct {
	io.Reader
	Name    string
	Total   uint64
	Current uint64
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)

	r.Current += uint64(n)
	fmt.Printf("\r%s\t进度 %.2f%%", r.Name, float64(r.Current*10000/r.Total)/100)
	return
}

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