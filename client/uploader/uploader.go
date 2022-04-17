package uploader

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

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
