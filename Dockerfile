FROM golang:1.19
USER root
# 设置环境变量
ENV GO111MODULE=on \
GOPROXY=https://goproxy.cn,direct \
CGO_ENABLED=0 \
GOOS=linux \
GOARCH=amd64

#移动到工作目录
WORKDIR /workspace/demo_go


ADD config.yaml .
ADD log.json .

COPY .env /app
COPY ./main-dev /app

EXPOSE 9093

CMD ["./main"]