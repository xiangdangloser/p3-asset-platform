# 阶段一：编译（Builder）
FROM golang:1.26-alpine AS builder

# 设置工作目录
WORKDIR /app

# ⭐ 核心修复：设置 Go 的国内代理，解决容器内下载依赖超时的问题
ENV GOPROXY=https://goproxy.cn,direct

# 安装编译工具（因为我们用的是纯 Go 的 SQLite 驱动，不需要 CGO，所以可以保持轻量）
# 提前拷贝依赖文件，利用 Docker 缓存层加速编译
COPY go.mod go.sum ./
RUN go mod download

# 拷贝所有源码
COPY . .

# 编译：开启 CGO_ENABLED=0 静态编译，关掉调试信息，让二进制体积最小
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o p3-asset-platform main.go

# 阶段二：运行（Runner）
# 使用轻量级的 Alpine Linux 作为运行环境
FROM alpine:latest

# 设置时区，保证数据库里的时间与本地一致
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

# 只把编译好的二进制文件从 Builder 阶段拷过来
COPY --from=builder /app/p3-asset-platform .

# 暴露 8080 端口
EXPOSE 8080

# 启动命令
CMD ["./p3-asset-platform"]