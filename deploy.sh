#!/bin/bash

echo "🚀 开始部署 X-UI 流量统计面板..."

# 创建数据目录
echo "📁 创建数据目录..."
mkdir -p data
chmod 755 data

# 停止并删除旧容器
echo "🛑 停止旧容器..."
docker-compose down

# 拉取最新镜像
echo "📥 拉取最新镜像..."
docker-compose pull

# 启动新容器
echo "🚀 启动新容器..."
docker-compose up -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
echo "🔍 检查服务状态..."
docker-compose ps

# 检查健康状态
echo "💚 检查健康状态..."
docker-compose exec x-ui-traffic wget --no-verbose --tries=1 --spider http://localhost:37022/health

if [ $? -eq 0 ]; then
    echo "✅ 部署成功！服务地址: http://$(hostname -I | awk '{print $1}'):37022"
else
    echo "❌ 部署失败，请检查日志:"
    docker-compose logs x-ui-traffic
fi 