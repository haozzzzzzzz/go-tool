#!/usr/bin/env bash
export GOROOT=/usr/local/go
export GOOS=windows
export GOARCH=amd64

# api tool
go build -o api.exe ../../api/main.go

# code tool
go build -o code.exe ../../code/main.go

## log formatter
#go build -o logfmt ${GOPATH}/src/github.com/haozzzzzzzz/go-rapid-development/tools/logfmt/main.go
#
## lambda build
#go build -o lamb ${GOPATH}/src/github.com/haozzzzzzzz/go-lambda/tools/lamb/main.go
#
## lambda deploy
#go build -o lamd ${GOPATH}/src/github.com/haozzzzzzzz/go-lambda/tools/lamd/main.go

echo finish