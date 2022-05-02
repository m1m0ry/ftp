```go
#交叉编译
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
```

```bash
#gofmt
find ./ -name "*.go" | xargs gofmt -w
```