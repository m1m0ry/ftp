package main

import (
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/m1m0ry/golang/ftp/server/common"
)

var configPath = flag.String("configPath", "./etc/config.json", "服务配置文件")
var confs = &common.ServiceConfig{}

// 列出文件信息
func listFiles(w http.ResponseWriter, request *http.Request) {
	files, err := ioutil.ReadDir(confs.StoreDir)
	if err != nil {
		fmt.Println("读文件夹失败", confs.StoreDir)
	}

	fileinfos := common.ListFileInfos{
		Files: []common.FileInfo{},
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		tmpFile := path.Join(confs.StoreDir, file.Name())
		fstate, err := os.Stat(tmpFile)
		if err != nil {
			fmt.Println("读取文件失败")
			continue
		}

		finfo := common.FileInfo{
			Filename: fstate.Name(),
			Filesize: fstate.Size(),
			Filetype: "normal",
		}

		fileinfos.Files = append(fileinfos.Files, finfo)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(fileinfos)
	if err != nil {
		fmt.Println("压缩文件列表失败")
		http.Error(w, "服务异常", http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)
}

// 处理upload逻辑
func upload(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("filename")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	hasher := &common.Hasher{
		Reader: file,
		Hash:   sha1.New(),
		Size:   0,
	}

	f, err := os.OpenFile(path.Join(confs.StoreDir, handler.Filename), os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
    
	io.Copy(f, hasher)
    
	if r.Header.Get("file-md5") != "" && r.Header.Get("file-md5") != hasher.Sum() {
		http.Error(w, "md5错误", http.StatusBadRequest)
		return
	}
}

// 下载文件
func download(w http.ResponseWriter, request *http.Request) {
	//文件名
	filename := request.FormValue("filename")

	//打开文件
	filePath := path.Join(confs.StoreDir, filename)
	file, err := os.Open(filePath)
	fstate, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("打开文件%s失败, err:%s\n", filePath, err)
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
	w.Header().Add("file-size", strconv.FormatUint(uint64(fstate.Size()), 10))

	//将文件写至responseBody
	_, err = io.Copy(w, hasher)
	if err != nil {
		http.Error(w, "文件下载失败", http.StatusInternalServerError)
		return
	}

	w.Header().Add("file-md5", hasher.Sum())

}

// 加载配置文件
func loadConfig(configPath string) {
	if !common.IsFile(configPath) {
		log.Panicf("config file %s is not exist", configPath)
	}

	buf, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Panicf("load config conf %s failed, err: %s\n", configPath, err)
	}

	err = json.Unmarshal(buf, confs)
	if err != nil {
		log.Panicf("decode config file %s failed, err: %s\n", configPath, err)
	}
}

func main() {
	//日志
	file, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("cant open log: ", err)
	}
	defer file.Close()

	log.SetOutput(file)
	log.SetPrefix("server main(): ")

	//服务器配置
	flag.Parse()
	loadConfig(*configPath)

	//路由设置
	http.HandleFunc("/list", listFiles)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/download", download)

	err = http.ListenAndServe(":"+strconv.Itoa(confs.Port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
