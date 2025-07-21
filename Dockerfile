# 构建阶段 - 前端
FROM node:18.19-alpine AS frontend-builder

WORKDIR /app/web

# 复制package文件并安装依赖
COPY web/package*.json ./
RUN npm ci --silent

# 复制前端源代码并构建
COPY web/ ./
RUN npm run build

# 构建阶段 - 后端
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app

# 安装系统依赖
RUN apk add --no-cache build-base musl-dev sqlite-dev

# 复制Go模块文件
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制后端源代码
COPY backend/ ./

# 静态编译后端（使用SQLite兼容性标签）
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -tags "sqlite_omit_load_extension" \
    -ldflags="-w -s -extldflags '-static'" \
    -o main .

# 最终运行镜像
FROM alpine:3.18

# 安装运行时依赖（静态编译后不需要sqlite）
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# 从构建阶段复制文件
COPY --from=frontend-builder /app/web/dist ./web/dist
COPY --from=backend-builder /app/main .

# 创建数据目录并设置权限
RUN mkdir -p /app/data && \
    chown -R appuser:appgroup /app && \
    chmod +x /app/main

# 切换到非root用户
USER appuser

# 设置环境变量
ENV DATABASE_PATH=/app/data/xtrafficdash.db
ENV X_UI_PASSWORD=admin123
ENV LISTEN_PORT=37022
ENV DEBUG_MODE=true
ENV LOG_LEVEL=debug

# 暴露端口
EXPOSE 37022

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:37022/health || exit 1

# 启动命令
CMD ["./main"] 