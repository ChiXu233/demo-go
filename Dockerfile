FROM alpine

WORKDIR /workspace/demo-go

COPY demo-go .

ADD config.yaml .
ADD log.json .

CMD ["./demo-go"]
