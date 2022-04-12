package common

//服务器地址
var BaseUrl string

// 文件元数据
type FileInfo struct {
	Filename    string  // 文件名
	Filesize    int64   // 文件大小
	Filetype    string  // 文件类型（分为普通文件和临时文件）
}

// 文件列表
type ListFileInfos struct {
	Files    []FileInfo
}