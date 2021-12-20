FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct

WORKDIR /build/zero

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
COPY app/usercenter/cmd/rpc/etc /app/etc
RUN go build -ldflags="-s -w" -o /app/usercenter app/usercenter/cmd/rpc/usercenter.go

FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates 
ENV TZ Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/usercenter /app/usercenter
COPY --from=builder /app/etc /app/etc

CMD ["./usercenter", "-f", "etc/usercenter.yaml"]
