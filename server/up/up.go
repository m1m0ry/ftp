package up

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/m1m0ry/golang/ftp/server/common"
	"os"
)

func Upload(uploader common.Uploader, uploadDir string) error {

	f, err := os.OpenFile(uploadDir, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()

	hasher := &common.Hasher{
		Reader: bytes.NewBuffer(uploader.Body),
		Hash:   sha1.New(),
		Size:   0,
	}

	buf := make([]byte, uploader.Size)
	hasher.Read(buf)
	_, err = f.WriteAt(buf, uploader.Offset)
	if err != nil {
		return err
	}

	if uploader.Md5 != hasher.Sum() {
		return errors.New("校验和不符")
	}

	return nil
}
