FROM golang:1.19
# 设置环境变量
ENV GO111MODULE=on \
GOPROXY=https://goproxy.cn,direct \
CGO_ENABLED=0 \
GOOS=linux \
GOARCH=amd64

#移动到工作目录
WORKDIR /workspace/demo-go

COPY . /demo-go

ADD config.yaml .
ADD log.json .


EXPOSE 9093

CMD ["./main"]