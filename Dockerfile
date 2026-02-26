# 第一阶段：构建Golang应用
FROM golang:1.24 AS builder

WORKDIR /app

# 设置GOPROXY
ENV GOPROXY=https://goproxy.cn,direct

# 拷贝go.mod和go.sum文件
COPY go.mod go.sum ./

# 预下载依赖
RUN go mod download

# 拷贝应用代码到镜像中
COPY . .

# 若前端 dist 不存在，创建最小占位页面，避免镜像构建失败
RUN mkdir -p dist && \
    [ -f dist/index.html ] || echo '<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Film Fusion</title></head><body><h1>Film Fusion</h1><p>前端资源未内置，请挂载完整 dist 目录。</p></body></html>' > dist/index.html

# 编译应用
RUN go build -o film-fusion

# 使用官方Alpine Linux镜像作为运行环境，由于它体积较小
FROM debian:stable-slim AS final

WORKDIR /app

# 安装tzdata包，添加时区数据
# RUN apk add bash tzdata sqlite

# 安装SQLite
RUN apt-get update && apt-get install -y sqlite3 bash tzdata ca-certificates


ENV VERSION=0.0.1
# 设置环境变量以配置时区
ENV TZ=Asia/Shanghai

# 从第一阶段复制编译好的应用到最终镜像
COPY --from=builder /app/film-fusion .
COPY --from=builder /app/dist ./dist

# 设置容器启动时运行的命令
CMD ["./film-fusion", "server"]
