#!/bin/sh

# 创建数据目录并设置权限
mkdir -p /app/data
chown -R 1001:1001 /app/data
chmod 755 /app/data

# 启动应用
exec ./main 