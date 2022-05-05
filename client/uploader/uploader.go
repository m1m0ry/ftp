package uploader

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/websocket"

	"github.com/m1m0ry/golang/ftp/client/common"
)

func UploadFile(filePath string) error {
	targetUrl := common.BaseUrl + "upload"

	if !common.IsFile(filePath) {
		fmt.Printf("filePath:%s is not exist", filePath)
		return errors.New(filePath + "文件不存在")
	}

	filename := filepath.Base(filePath)
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("filename", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}

	//打开文件句柄操作
	fh, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("error opening filePath: %s\n", filePath)
		return err
	}

	hasher := &common.Hasher{
		Reader: fh,
		Hash:   sha1.New(),
		Size:   0,
	}

	//iocopy
	_, err = io.Copy(fileWriter, hasher)
	if err != nil {
		return err
	}
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	//resp, err := http.Post(targetUrl, contentType, bodyBuf)

	request, err := http.NewRequest(http.MethodPost, targetUrl, bodyBuf)
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", contentType)
	request.Header.Add("file-md5", hasher.Sum())
	fmt.Println(hasher.Sum())

	resp, err := http.DefaultClient.Do(request) // enter 键
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("%s文件上传失败\n", filename)
		return errors.New("上传文件失败")
	}

	fmt.Printf("上传文件%s成功\n", filename)
	return nil
}

func Upload(filePath string, url string) error {
	infoUrl := common.BaseUrl + "info"

	if !common.IsFile(filePath) {
		fmt.Printf("filePath:%s is not exist", filePath)
		return errors.New(filePath + "文件不存在")
	}

	filename := filepath.Base(filePath)

	//打开文件句柄操作
	fstat, _ := os.Stat(filePath)

	finfo := common.FileInfo{
		Filename: fstat.Name(),
		Filesize: fstat.Size(),
	}

	data, err := json.Marshal(finfo)
	if err != nil {
		fmt.Println("json.marshal failed, err:", err)
		return err
	}

	_, err = http.Post(infoUrl, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("err:", err)
		return err
	}

	var maxRun int

	switch g := int64(1024 * 1024 * 1024); {
	case finfo.Filesize < g:
		maxRun = 4
	case finfo.Filesize < 4*g:
		maxRun = 16
	case finfo.Filesize < 16*g:
		maxRun = 32
	case finfo.Filesize < 64*g:
		maxRun = 64
	default:
		maxRun = 0
	}

	c, _, err := websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		log.Panicln(err)
	}

	ch := make(chan error, maxRun)

	size := fstat.Size()
	part := (size / int64(maxRun))

	for offset := int64(0); offset < size; offset = offset + part {
		if offset > size-part {
			part = size - offset
		}
		go UploadPart(offset, part, filePath, c, ch)
	}

	for i := 0; i < maxRun; i++ {
		if err := <-ch; err != nil {
			log.Println(err)
		}
	}

	fmt.Printf("上传文件%s成功\n", filename)
	return nil
}

func UploadPart(offset int64, part int64, filePath string, c *websocket.Conn, ch chan error) {
	filename := filepath.Base(filePath)
	fh, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("error opening filePath: %s\n", filePath)
		ch <- err
		return
	}

	fh.Seek(offset, io.SeekStart)

	hasher := &common.Hasher{
		Reader: fh,
		Hash:   sha1.New(),
		Size:   0,
	}

	buf := make([]byte, part)
	hasher.Read(buf)

	uploader := common.Uploader{
		Body:   buf,
		Name:   filename,
		Size:   part,
		Offset: offset,
		Md5:    hasher.Sum(),
	}

	err = c.WriteJSON(uploader)
	ch <- err
}
