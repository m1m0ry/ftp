package main

import (
	"encoding/json"
	"flag"
	"fmt"
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

// 加载配置文件
func loadConfig(configPath string) () {
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
        log.Fatal("cant open log: ",err)
    }
    defer file.Close()

    log.SetOutput(file)
    log.SetPrefix("server main(): ")

    //服务器配置
    flag.Parse()
    loadConfig(*configPath)

    //路由设置
    http.HandleFunc("/listFiles", listFiles)

    err = http.ListenAndServe(":"+strconv.Itoa(confs.Port), nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}