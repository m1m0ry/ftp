package down

import (
	"crypto/sha1"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/m1m0ry/golang/ftp/server/common"
)

//下载所选文件的对应部分
func Download(filePath string, offset uint64, size uint64, w http.ResponseWriter) {
	filename := filepath.Base(filePath)
	file, err := os.Open(filePath)
	fstate, _ := os.Stat(filePath)
	if err != nil {
		log.Printf("打开文件%s失败, err:%s\n", filePath, err)
		http.Error(w, "文件打开失败", http.StatusBadRequest)
		return
	}
	//结束后关闭文件
	defer file.Close()

	hasher := &common.Hasher{
		Reader: file,
		Hash:   sha1.New(),
		Size:   0,
	}

	//设置响应的header头
	w.Header().Add("content-type", "application/octet-stream")
	w.Header().Add("content-disposition", "attachment; filename=\""+filename+"\"")

	if offset == size && size == 0 {
		w.Header().Add("file-size", strconv.FormatUint(uint64(fstate.Size()), 10))

		//将文件写至responseBody
		_, err = io.Copy(w, hasher)
	}else{
		w.Header().Add("file-size", strconv.FormatUint(size, 10))

		buf:=make([]byte,size)
		hasher.Read(buf)
		_, err = w.Write(buf)
	}

	
	if err != nil {
		http.Error(w, "文件下载失败", http.StatusInternalServerError)
		return
	}
	w.Header().Add("file-md5", hasher.Sum())
}
