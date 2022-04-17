# ftp

## v 0.05

### todo
- 命令行操作
- 服务器配置
- 列出服务器文件目录
- 简单日志

### ~ 2day
- day 1：

### then
- 配置服务器，索引/解析文件目录 json
- 日志 log
- 命令行 flag
- 文件列表显示 gotable

### bash
go run main.go

go run main.go --action list

## v 0.07

### todo
- 文件上传，下载，校验
- 多文件上传

### ~ 3day
- day 1：上传,多文件上传
- day 2；下载,校验和

### then
- 校验和,上传前计算，服务器端计算, 下载后计算
- 多文件，并发
### bash
go run main.go --action upload -uploadFilepaths {path}
go run main.go --action download -downloadFilenames {path}