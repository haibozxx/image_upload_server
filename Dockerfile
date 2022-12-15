FROM golang:1.18.9-alpine3.17 as builder

WORKDIR /app

ENV GOPROXY="https://goproxy.cn,direct"
ENV GO111MODULE="on"

COPY . .

RUN go mod tidy && GOOS=linux GOARCH=amd64 go build -o image_uploader .


FROM alpine:latest as prod

WORKDIR /app

COPY --from=builder /app/image_uploader .

CMD ["./image_uploader"]
