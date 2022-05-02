package downloader

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/m1m0ry/golang/ftp/client/common"
	"github.com/m1m0ry/golang/ftp/client/db"
)

// DownloadFile 单个文件的下载
func DownloadFile(filename string, downloadDir string) error {
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
		Hash:   sha1.New(),
		Size:   0,
	}

	var size uint64 = 1
	if r.Header.Get("file-size") != "" {
		size, _ = strconv.ParseUint(r.Header.Get("file-size"), 10, 64)
	}

	//size,_:=strconv.ParseUint(r.Header.Get("file-size"),10,64)

	reader := &common.Reader{
		Reader: hasher,
		Name:   filename,
		Total:  size,
	}

	_, err = io.Copy(f, reader)
	if err != nil {
		return err
	}

	if r.Header.Get("file-md5") != "" && r.Header.Get("file-md5") != hasher.Sum() {
		fmt.Println("文件下载错误")
		return nil
	}
	fmt.Printf("\n%s 文件下载成功，保存路径：%s\n", filename, filePath)
	return nil
}

func Download(filename string, downloadDir string) error {
	if !common.IsDir(downloadDir) {
		fmt.Printf("指定下载路径：%s 不存在\n", downloadDir)
		return errors.New("指定下载路径不存在")
	}
	filePath := path.Join(downloadDir, filename)
	infoUrl := common.BaseUrl + "info?filename=" + filename

	fileInfo := db.Init(filePath)

	var offset int64
	var size int64
	if fileInfo.Old {
		offset = fileInfo.IsDone(filePath)
		var info common.FileInfo
		err := json.Unmarshal(fileInfo.Get(filePath), &info)
		if err != nil {
			log.Fatal("unmarshal err: ", err)
		}
		size = info.Filesize
	} else {
		r, err := http.Get(infoUrl)
		if err != nil {
			fmt.Println("获取文件信息失败", err.Error())
			return err
		}
		defer r.Body.Close()
		var info common.FileInfo
		err = json.NewDecoder(r.Body).Decode(&info)
		if err != nil {
			log.Fatal("unmarshal err: ", err)
		}
		fileInfo.Post(filePath, info)
		size = info.Filesize
		offset = 0
	}

	//v 0.01 单线程下载
	ack := make(chan bool)
	go func() { ack <- true }()
	for off := offset; <-ack && off < size; {

		fileInfo.Put(filePath, off)
		part := common.MaxPart
		if off > size-common.MaxPart {
			part = size - off
		}

		go downloadPart(filename, filePath, off, part, ack)
		off = off + part
	}

	return nil
}

func downloadPart(filename string, filePath string, offset int64, size int64, ack chan bool) error {

	targetUrl := common.BaseUrl + "download?filename=" + filename

	//请求

	request, err := http.NewRequest(http.MethodGet, targetUrl, nil)
	if err != nil {
		return err
	}
	request.Header.Add("offset", strconv.FormatInt(offset, 10))
	request.Header.Add("size", strconv.FormatInt(size, 10))

	//响应

	resp, err := http.DefaultClient.Do(request) // enter 键
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//写入

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf(err.Error())
		return err
	}
	defer f.Close()

	hasher := &common.Hasher{
		Reader: resp.Body,
		Hash:   sha1.New(),
		Size:   0,
	}

	buf := make([]byte, size)
	hasher.Read(buf)
	_, err = f.WriteAt(buf, offset)
	if err != nil {
		return err
	}

	if resp.Header.Get("file-md5") != "" && resp.Header.Get("file-md5") != hasher.Sum() {
		fmt.Println("文件下载错误")
		return nil
	}

	ack <- true

	return nil
}
