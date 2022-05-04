package db

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/m1m0ry/golang/ftp/server/common"
)

func TestInitFileStore(t *testing.T) {
	store, err := InitSqliteStore("test")

	if err != nil {
		t.Error("err info: ", err)
	}

	if _, ok := store.(Store); ok != true {
		t.Errorf("bad store : %q", store)
	}
	os.Remove("test")
}

func TestPost(t *testing.T) {
	store, _ := InitSqliteStore("test")
	var info common.FileInfo
	err := store.Post("test", info)
	if err != nil {
		t.Errorf("err info: %q", err)
	}
	os.Remove("test")
}

func TestGet(t *testing.T) {
	store, _ := InitSqliteStore("test")
	want := common.FileInfo{
		Filename: "test",
		Filesize: 1024,
		Offset:   16,
	}
	store.Post("test", want)
	var get common.FileInfo
	json.Unmarshal(store.Get("test"), &get)
	if want != get {
		t.Error("want ", want, " get ", get, " not equal")
	}
	os.Remove("test")
}

func TestPut(t *testing.T) {
	store, _ := InitSqliteStore("test")
	store.Put("test", 1024)
	os.Remove("test")
}

func TestIsDone(t *testing.T) {
	store, _ := InitSqliteStore("test")
	store.Put("test", 1024)
	if tmp := store.IsDone("test"); tmp != 1024 {
		t.Error("want 1024 get ", tmp)
	}
	os.Remove("test")
}

func TestDelete(t *testing.T) {
	store, _ := InitSqliteStore("test")
	want := common.FileInfo{}
	store.Post("test", want)
	store.Delete("test")
	var get common.FileInfo
	json.Unmarshal(store.Get("test"), &get)
	if want != get {
		t.Error("want ", want, " get ", get, " not equal")
	}
	os.Remove("test")
}
