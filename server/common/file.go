package common

// 文件元数据
type FileInfo struct {
	Filename string `json:"name"`
	Filesize int64  `json:"size"`
	Offset   int64  `json:"offset"`
}

// 文件列表
type ListFileInfos struct {
	Files []FileInfo
}
