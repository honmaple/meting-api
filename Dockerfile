FROM golang:1.19-alpine AS builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache gcc musl-dev

WORKDIR /src

COPY go.mod go.sum /src
RUN GOPROXY=https://proxy.golang.com.cn,direct go mod download

COPY main.go /src/main.go
COPY music /src/music
COPY internal /src/internal
RUN go build .

FROM alpine:latest

WORKDIR /opt/meting-api

COPY --from=builder /src/meting-api /usr/bin

CMD ["/usr/bin/meting-api", "-D"]