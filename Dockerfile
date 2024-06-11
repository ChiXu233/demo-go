FROM alpine

WORKDIR /workspace/demo-go

COPY demo-go .

ADD config.yaml .
ADD log.json .

EXPOSE 9093

CMD ["./demo-go"]
