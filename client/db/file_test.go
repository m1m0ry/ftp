package db

import (
	"encoding/json"
	"testing"

	"github.com/m1m0ry/golang/ftp/client/common"
)

func TestInitFileStore(t *testing.T) {
	store, err := InitFileStore("test")

	if err != nil {
		t.Logf("err info: %q", err)
	}

	if _, ok := store.(Store); ok != true {
		t.Errorf("bad store : %q", store)
	}
}

func TestPost(t *testing.T) {
	store, _ := InitFileStore("test")
	var info common.FileInfo
	err := store.Post("test", info)
	if err != nil {
		t.Errorf("err info: %q", err)
	}
}

func TestGet(t *testing.T) {
	store, _ := InitFileStore("test")
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
}

func TestPut(t *testing.T) {
	store, _ := InitFileStore("test")
	store.Put("test", 1024)
}

func TestIsDone(t *testing.T) {
	store, _ := InitFileStore("test")
	store.Put("test", 1024)
	if tmp := store.IsDone("test"); tmp != 1024 {
		t.Error("want 1024 get ", tmp)
	}
}

func TestDelete(t *testing.T) {
	store, _ := InitFileStore("test")
	store.Delete("test")
	want := common.FileInfo{}
	var get common.FileInfo
	json.Unmarshal(store.Get("test"), &get)
	if want != get {
		t.Error("want ", want, " get ", get, " not equal")
	}
}
