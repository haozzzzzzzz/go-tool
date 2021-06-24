#!/usr/bin/env bash
export GOROOT=/usr/local/go
export GOOS=darwin
export GOARCH=amd64

# api tool
build_time=`date +"%Y-%m-%d %H:%M:%S %Z"`
#echo $build_time
go build -ldflags="-X 'main.BuildTime=${build_time}'" -o api ../../api/main.go

# ws doc tool
go build -ldflags="-X 'main.BuildTime=${build_time}'" -o ws ../../ws/main.go

# code tool
#go build -o code ../../code/main.go

## log formatter
#go build -o logfmt ${GOPATH}/src/github.com/haozzzzzzzz/go-rapid-development/v2/tools/logfmt/main.go
#
## lambda build
#go build -o lamb ${GOPATH}/src/github.com/haozzzzzzzz/go-lambda/tools/lamb/main.go
#
## lambda deploy
#go build -o lamd ${GOPATH}/src/github.com/haozzzzzzzz/go-lambda/tools/lamd/main.go

echo finish
