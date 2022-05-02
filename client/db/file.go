package db

import (
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/m1m0ry/golang/ftp/client/common"
)

type Store interface {
	Close(filePath string)
	Get(filePath string) []byte
	Post(filePath string, fileinfo interface{}) (err error)
	Put(filePath string, offset int64)
	IsDone(filePath string) int64
	Delete(filePath string)
}

type FileStore struct {
	infoPath   string
	offsetPath string
	Old        bool
}

func Init(filePath string) *FileStore {
	info := &FileStore{
		infoPath:   filePath + ".downloading",
		offsetPath: filePath + ".2downloading",
	}
	if common.IsFile(info.infoPath) {
		info.Old = true
	}
	return info
}

func (f FileStore) Get(filePath string) []byte {
	content, _ := ioutil.ReadFile(f.infoPath)
	return content
}

func (f FileStore) Post(filePath string, fileinfo interface{}) (err error) {
	content, err := json.Marshal(fileinfo)
	if err != nil {
		return
	}
	return ioutil.WriteFile(f.infoPath, content, 0644)
}

func (f FileStore) Put(filePath string, offset int64) {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, uint64(offset))
	err := ioutil.WriteFile(f.offsetPath, bs, 0644)
	if err != nil {
		log.Panicln(err)
	}
}

func (f FileStore) IsDone(filePath string) int64 {
	if !common.IsFile(f.offsetPath) {
		return 0
	}
	content, err := ioutil.ReadFile(f.offsetPath)
	if err != nil {
		log.Panicln(err)
	}
	return int64(binary.LittleEndian.Uint64(content))
}

func (f FileStore) Delete(filePath string) {
	err := os.Remove(f.infoPath)
	if err != nil {
		log.Println("file remove Error!")
		log.Printf("%s", err)
	} else {
		log.Print("file remove OK!")
	}
	err = os.Remove(f.offsetPath)
	if err != nil {
		log.Println("file remove Error!")
		log.Printf("%s", err)
	} else {
		log.Print("file remove OK!")
	}
}
