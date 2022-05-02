package db

import (
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Store interface {
	Close(filePath string)
	Get(filePath string) string
	Post(filePath string, fileinfo interface{}) (err error)
	Put(filePath string, offset int64)
	IsDone(filePath string) int64
	Delete(filePath string)
}

type FileStore struct {
	file *os.File
	filePath string
	Old bool
}

func Init(name string) *FileStore{
	return &FileStore{
		filePath: name,
	}
}

func (f FileStore) Get(filePath string) string {
	file := f.isOpen(filePath)
	content, _ := ioutil.ReadAll(file)
	return string(content)
}

func (f FileStore) Post(filePath string, fileinfo interface{}) (err error) {
	content, err := json.Marshal(fileinfo)
	if err != nil {
		return
	}
	return ioutil.WriteFile(filePath, content, 0644)
}

func (f FileStore) Put(filePath string, offset int64) {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, uint64(offset))
	err := ioutil.WriteFile(filePath+".2downloading", bs, 0644)
	if err != nil {
		log.Panicln(err)
	}
}

func (f FileStore) IsDone(filePath string) int64 {
	content, err := ioutil.ReadFile(filePath + ".2downloading")
	if err != nil {
		log.Panicln(err)
	}
	return int64(binary.LittleEndian.Uint64(content))
}

func (f FileStore) Delete(filePath string) {
	f.Close(filePath)
	err := os.Remove(filePath + ".downloading")
	if err != nil {
		log.Println("file remove Error!")
		log.Printf("%s", err)
	} else {
		log.Print("file remove OK!")
	}
	err = os.Remove(filePath + ".2downloading")
	if err != nil {
		log.Println("file remove Error!")
		log.Printf("%s", err)
	} else {
		log.Print("file remove OK!")
	}
}

func (f FileStore) Close(filePath string) {
	file := f.isOpen(filePath)
	file.Close()
}

func (f FileStore) isOpen(filePath string) *os.File {
	if f.file == nil {
		f.file, _ = os.OpenFile(filePath+".downloading", os.O_WRONLY|os.O_CREATE, 0666)
		return f.file
	} else {
		f.Old=true
		return f.file
	}
}
