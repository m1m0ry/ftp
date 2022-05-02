package common

import (
	"fmt"
	"io"
)

//服务器地址
var BaseUrl string

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