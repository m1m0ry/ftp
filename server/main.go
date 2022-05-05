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

	"github.com/gorilla/websocket"

	"github.com/m1m0ry/golang/ftp/server/common"
	"github.com/m1m0ry/golang/ftp/server/db"
	"github.com/m1m0ry/golang/ftp/server/down"
	"github.com/m1m0ry/golang/ftp/server/up"
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

var upgrader = websocket.Upgrader{} // use default options

func wsUpload(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		var uploader common.Uploader
		err := c.ReadJSON(uploader)
		if err != nil {
			log.Println("read:", err)
			break
		}
		go func() {
			err = up.Upload(uploader, path.Join(confs.StoreDir, uploader.Name))
			if err != nil {
				err = c.WriteJSON(err)
				if err != nil {
					log.Println("write:", err)
				}
			}
		}()
	}
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
	offset, _ := strconv.ParseUint(request.Header.Get("offset"), 10, 64)
	size, _ := strconv.ParseUint(request.Header.Get("size"), 10, 64)
	down.Download(filePath, int64(offset), int64(size), w)
}

// 文件元数据
func info(w http.ResponseWriter, request *http.Request) {
	if request.Method == "Get" {
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
			Status:   true,
			Host:     request.Host,
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(finfo)
		if err != nil {
			fmt.Println("压缩文件列表失败")
			http.Error(w, "服务异常", http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
	} else {
		var finfo common.FileInfo
		err := json.NewDecoder(request.Body).Decode(&finfo)
		if err != nil {
			fmt.Println("文件信息读取失败")
			http.Error(w, "信息读取失败", http.StatusBadRequest)
		}
		if finfo.Filesize > 64*1024*1024*1024 {
			http.Error(w, "文件过大, 暂不支持超过64G的文件", http.StatusPreconditionFailed)
		}
		filePath := path.Join(confs.StoreDir, finfo.Filename)
		info, _ := db.InitSqliteStore(filePath)
		info.Post(filePath, finfo)
	}
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
	http.HandleFunc("/wsupload", wsUpload)
	http.HandleFunc("/download", download)
	http.HandleFunc("/info", info)

	err = http.ListenAndServe(":"+strconv.Itoa(confs.Port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
