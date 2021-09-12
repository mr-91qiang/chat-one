#!/bin/bash
echo "开始编译"

export CGO_ENABLED=0
go mod tidy
go build
if test -e ./char
then
  echo "编译完成"
else
  ls
  echo "编译失败"
  exit 0
fi
docker build  --no-cache   -t char:0.1 .
docker rm -f char
docker run -it -d -p5900:5900 --name char  char:0.1
echo "发布成功"