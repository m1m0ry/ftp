package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/liushuochen/gotable"
	"github.com/m1m0ry/golang/ftp/client/common"
)

// 定义命令行参数对应的变量
var serverIP = flag.String("serverIP", "127.0.0.1", "服务IP")
var serverPort = flag.Int("serverPort", 10808, "服务端口")
var action = flag.String("action", "", "upload, download or list")
var uploadFile = flag.String("uploadFile", "", "上传文件路径,多个文件路径用空格相隔")
var downloadFile = flag.String("downloadFile", "", "下载文件名")
var downloadDir = flag.String("downloadDir", "/download", "下载路径，默认当前目录")

// listFiles 列出文件列表
func listFiles() {
	targetUrl := common.BaseUrl + "listFiles"

	req, _ := http.NewRequest("GET", targetUrl, nil)
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		fmt.Println("获取文件列表信息失败", err.Error())
		return
	}
	defer resp.Body.Close()

	var fileinfos common.ListFileInfos
	err = json.NewDecoder(resp.Body).Decode(&fileinfos)
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
		err := table.AddRow([]string{fileinfo.Filename,strconv.FormatInt(fileinfo.Filesize,10)})
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
		//uploadFiles(*uploadFile)
	case "download":
		// 下载文件
		//downloadFiles(*downloadFile, *downloadDir)
	case "list":
		// 列出文件
		listFiles()
	default:
		fmt.Printf("unknow action: %s\n", *action)
		os.Exit(-1)
	}
}
