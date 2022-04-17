package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/liushuochen/gotable"
	"github.com/m1m0ry/golang/ftp/client/common"
	"github.com/m1m0ry/golang/ftp/client/downloader"
	"github.com/m1m0ry/golang/ftp/client/uploader"
)

// 定义全局变量
var globalWait sync.WaitGroup

// 定义命令行参数对应的变量
var serverIP = flag.String("ip", "127.0.0.1", "服务IP")
var serverPort = flag.Int("port", 10808, "服务端口")
var action = flag.String("action", "", "upload, download list or host")
var uploadFilepaths = flag.String("uploadFilepaths", "", "上传文件路径,多个文件路径用空格相隔")
var downloadFilenames = flag.String("downloadFilenames", "", "下载文件名")
var downloadDir = flag.String("downloadDir", "download", "下载路径，默认当前目录")

// 下载文件
func downloadFile(filename string, downloadDir string) {
	defer globalWait.Done()

	err := downloader.DownloadFile(filename, downloadDir)
	if err != nil {
		fmt.Printf("%s文件下载失败", filename)
	}
}

func downloadFiles(filePaths string, downloadDir string) {
	if !common.IsDir(downloadDir) {
		fmt.Println("路径不存在", downloadDir)
		os.Exit(-1)
	}

	files := strings.Split(filePaths, " ")
	for _, file := range files {
		globalWait.Add(1)
		go downloadFile(file, downloadDir)
	}
	globalWait.Wait()
}

// 上传文件
func uploadFile(uploadFilepath string) {

	defer globalWait.Done()

	err := uploader.UploadFile(uploadFilepath)

	if err != nil {
		fmt.Printf("上传%s文件失败\n", uploadFilepath)
	}
}

func uploadFiles(uploadFilepaths string) {

	if uploadFilepaths == "" {
		fmt.Println("文件路径为空，请检查命令格式是否正确")
		return
	}

	// 以空格方式分割要上传的文件
	files := strings.Split(uploadFilepaths, " ")
	for _, file := range files {
		globalWait.Add(1)
		go uploadFile(file)
	}
	globalWait.Wait()
}

// listFiles 列出文件列表
func listFiles() {
	targetUrl := common.BaseUrl + "list"

	r, err := http.Get(targetUrl)
	if err != nil {
		fmt.Println("获取文件列表信息失败", err.Error())
		return
	}
	defer r.Body.Close()
	content, err := ioutil.ReadAll(r.Body)
	fmt.Printf("%s",content)
	var fileinfos common.ListFileInfos
	err = json.NewDecoder(r.Body).Decode(&fileinfos)
	if err != nil {
		fmt.Println("获取文件列表信息失败")
		return
	}

	table, err := gotable.Create("name", "size")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, fileinfo := range fileinfos.Files {
		err := table.AddRow([]string{fileinfo.Filename, strconv.FormatInt(fileinfo.Filesize, 10)})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	//边框
	//table.CloseBorder()
	fmt.Println(table)
}

func main() {

	// 解析传入的参数
	flag.Parse()

	// 设置基础请求URL值
	common.BaseUrl = fmt.Sprintf("http://%s:%d/", *serverIP, *serverPort)

	switch *action {
	case "upload":
		// 上传文件
		uploadFiles(*uploadFilepaths)
	case "download":
		// 下载文件
		downloadFiles(*downloadFilenames, *downloadDir)
	case "list":
		// 列出文件
		listFiles()
	//case "host":
		// 网络发现
		//host()
	//default:
		fmt.Printf("unknow action: %s\n", *action)
		os.Exit(-1)
	}
}
