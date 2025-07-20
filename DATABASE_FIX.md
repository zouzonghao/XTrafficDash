# 数据库权限修复说明

## 问题描述

在数据库迁移后，出现了以下错误：
```
"存储流量数据失败: 记录原始请求失败: attempt to write a readonly database"
```

## 问题原因

1. **进程冲突**：有多个后端进程同时访问数据库文件
2. **文件锁定**：SQLite数据库被某个进程锁定为只读模式
3. **权限问题**：数据库文件权限设置不当

## 解决方案

### 1. 停止冲突进程
```bash
# 查看使用数据库的进程
lsof xtrafficdash.db

# 停止相关进程
kill <PID>
```

### 2. 修复文件权限
```bash
# 设置正确的文件权限
chmod 644 xtrafficdash.db
```

### 3. 验证数据库可写性
```bash
# 测试数据库写入
sqlite3 xtrafficdash.db "INSERT INTO raw_requests (service_id, client_ip, request_body) VALUES (1, '127.0.0.1', 'test');"

# 清理测试数据
sqlite3 xtrafficdash.db "DELETE FROM raw_requests WHERE client_ip = '127.0.0.1' AND request_body = 'test';"
```

### 4. 重启后端服务
```bash
cd backend
go run main.go &
```

## 验证结果

### 修复前
```
{"level":"error","msg":"存储流量数据失败: 记录原始请求失败: attempt to write a readonly database"}
```

### 修复后
```
{"level":"info","msg":"流量数据已存储到数据库"}
{"success":true,"message":"服务正常运行","data":{"database":"connected"}}
```

## 预防措施

1. **单一进程**：确保同一时间只有一个后端进程运行
2. **正确停止**：使用 `Ctrl+C` 或 `kill` 命令正确停止服务
3. **权限检查**：定期检查数据库文件权限
4. **进程监控**：使用 `lsof` 命令监控数据库文件使用情况

## 修复完成确认

- ✅ 停止冲突进程
- ✅ 修复文件权限
- ✅ 验证数据库可写性
- ✅ 重启后端服务
- ✅ 流量数据正常存储
- ✅ 健康检查接口正常

修复时间：2025-07-20 15:27:35 