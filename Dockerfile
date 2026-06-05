# 第一阶段：用标准的 Go 环境进行编译
FROM golang:1.26-alpine AS builder

WORKDIR /app

# 复制依赖并下载
COPY go.mod go.sum ./
RUN go mod tidy

# 复制源码
COPY . .

# Build
RUN apk add --no-cache --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main/ nodejs npm

RUN cd FrontEnd && npm install && npm run build && cd ..

RUN CGO_ENABLED=0 GOOS=linux go build -o mahjong-server .

# 运行
FROM alpine:latest

WORKDIR /app

# 从上一个阶段把编译好的静态二进制文件复制过来
COPY --from=builder /app/mahjong-server ./

EXPOSE 8080

CMD ["./mahjong-server"]











