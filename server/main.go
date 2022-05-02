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
	"github.com/m1m0ry/golang/ftp/server/down"
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
			Offset:   0,
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
	filePath := path.Join(confs.StoreDir, filename)
	fmt.Println(request)
	offset, _ := strconv.ParseUint(request.Header.Get("offset"), 10, 64)
	size, _ := strconv.ParseUint(request.Header.Get("size"), 10, 64)
	down.Download(filePath, int64(offset), int64(size), w)
}

// 文件元数据
func info(w http.ResponseWriter, request *http.Request) {

	filename := request.FormValue("filename")
	filePath := path.Join(confs.StoreDir, filename)

	fstate, err := os.Stat(filePath)
	if err != nil {
		fmt.Println("读取文件失败")
	}

	finfo := common.FileInfo{
		Filename: fstate.Name(),
		Filesize: fstate.Size(),
		Offset:   0,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(finfo)
	if err != nil {
		fmt.Println("压缩文件列表失败")
		http.Error(w, "服务异常", http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)
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
	http.HandleFunc("/info", info)

	err = http.ListenAndServe(":"+strconv.Itoa(confs.Port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
