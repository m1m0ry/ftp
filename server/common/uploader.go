package common

type Uploader struct {
	Body   []byte `json:"body"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Offset int64  `json:"offset"`
	Md5    string `json:"md5"`
}
