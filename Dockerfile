FROM alpine

WORKDIR /workspace/demo-go

COPY demo-go .

CMD ["./demo-go"]
