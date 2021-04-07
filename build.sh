mkdir -p dist

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
go build -ldflags '-w' -o dist/logv logv/log_viewer.go
go build -ldflags '-w' -o dist/sskey sskey/sskey.go
