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
```go
go run main.go
go run main.go --action list
```

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
```go
go run main.go --action upload -uploadFilepaths {path}
go run main.go --action download -downloadFilenames {path}
```

## v 0.08
### todo
- fix bug
- clean bad code
### ~ 2day
- day 1：

### then
- 错误处理，路径错误，无所选文件
- 错误日志
- 文件路径处理
- 空哈希错误 da39a3ee5e6b4b0d3255bfef95601890afd80709

## v 0.09

### todo
- 分片上传/下载
- rowdata 数据库