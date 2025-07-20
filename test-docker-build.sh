#!/bin/bash

echo "🧪 开始测试Docker构建..."

# 清理之前的构建
echo "🧹 清理之前的构建..."
docker rmi xtrafficdash:test 2>/dev/null || true

# 构建Docker镜像
echo "🔨 构建Docker镜像..."
docker build -t xtrafficdash:test .

if [ $? -eq 0 ]; then
    echo "✅ Docker构建成功！"
    
    # 测试运行容器
    echo "🚀 测试运行容器..."
    docker run -d --name xtrafficdash-test -p 37022:37022 xtrafficdash:test
    
    # 等待服务启动
    echo "⏳ 等待服务启动..."
    sleep 5
    
    # 测试健康检查
    echo "🏥 测试健康检查..."
    if curl -f http://localhost:37022/health >/dev/null 2>&1; then
        echo "✅ 健康检查通过！"
    else
        echo "❌ 健康检查失败"
    fi
    
    # 清理测试容器
    echo "🧹 清理测试容器..."
    docker stop xtrafficdash-test 2>/dev/null || true
    docker rm xtrafficdash-test 2>/dev/null || true
    
    echo "🎉 所有测试完成！"
else
    echo "❌ Docker构建失败！"
    exit 1
fi 