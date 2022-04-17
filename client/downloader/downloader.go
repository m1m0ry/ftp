package downloader

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/m1m0ry/golang/ftp/client/common"
)

// DownloadFile 单个文件的下载
func DownloadFile(filename string, downloadDir string) (error){
	if !common.IsDir(downloadDir) {
		fmt.Printf("指定下载路径：%s 不存在\n", downloadDir)
		return errors.New("指定下载路径不存在")
	}

	targetUrl := common.BaseUrl + "download?filename=" + filename
	r, err := http.Get(targetUrl)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer r.Body.Close()

	filePath := path.Join(downloadDir, filename)
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf(err.Error())
		return err
	}
	defer f.Close()

	hasher := &common.Hasher{
		Reader: r.Body,
		Hash: sha1.New(),
		Size: 0,
	}

	size,_:=strconv.ParseUint(r.Header.Get("file-size"),10,64)

	reader := &common.Reader{
		Reader: hasher,
		Name: filename,
		Total: size,
	}

	if(r.Header.Get("file-md5")!=""&&r.Header.Get("file-md5")!=hasher.Sum()){
		fmt.Println("文件下载错误")
        return nil
    }

	fmt.Println(hasher.Sum())

	_, err = io.Copy(f, reader)
	if err != nil {
		return err
	}
	fmt.Printf("\n%s 文件下载成功，保存路径：%s\n", filename, filePath)
	return nil
}