# 数据库迁移说明

## 迁移概述

由于项目名称从 `x-ui-panel` 更改为 `XTrafficDash`，数据库文件名也相应从 `xui_traffic.db` 更改为 `xtrafficdash.db`。

## 迁移步骤

### 1. 备份原数据库
```bash
cd backend
cp xui_traffic.db xui_traffic.db.backup
```

### 2. 重命名数据库文件
```bash
mv xui_traffic.db xtrafficdash.db
```

### 3. 验证迁移结果
```bash
# 检查数据库表结构
sqlite3 xtrafficdash.db "SELECT name FROM sqlite_master WHERE type='table';"

# 检查数据完整性
sqlite3 xtrafficdash.db "SELECT COUNT(*) FROM services;"
sqlite3 xtrafficdash.db "SELECT COUNT(*) FROM client_traffics;"
sqlite3 xtrafficdash.db "SELECT COUNT(*) FROM inbound_traffics;"
```

## 迁移验证

### 数据库表结构
- ✅ services - 服务表
- ✅ inbound_traffics - 入站流量表
- ✅ client_traffics - 客户端流量表
- ✅ inbound_traffic_history - 入站流量历史表
- ✅ client_traffic_history - 客户端流量历史表
- ✅ raw_requests - 原始请求表

### 数据完整性
- ✅ 服务数据：3个服务记录
- ✅ 客户端数据：8个用户记录
- ✅ 入站流量数据：7个端口记录
- ✅ 历史数据：保留完整的历史记录

## 配置更新

### 环境变量
```bash
# 旧配置
DATABASE_PATH=xui_traffic.db

# 新配置
DATABASE_PATH=xtrafficdash.db
```

### Docker配置
```yaml
# 旧配置
environment:
  - DATABASE_PATH=/app/data/xui_traffic.db

# 新配置
environment:
  - DATABASE_PATH=/app/data/xtrafficdash.db
```

## 注意事项

1. **备份文件**：原数据库已备份为 `xui_traffic.db.backup`
2. **自动迁移**：数据库表结构会自动迁移，无需手动干预
3. **数据保留**：所有历史数据都已保留
4. **服务兼容**：后端服务会自动使用新的数据库文件名

## 回滚方案

如需回滚到原数据库：
```bash
cd backend
rm xtrafficdash.db
mv xui_traffic.db.backup xui_traffic.db
# 修改环境变量 DATABASE_PATH=xui_traffic.db
```

## 迁移完成确认

- ✅ 数据库文件重命名完成
- ✅ 数据完整性验证通过
- ✅ 后端服务启动正常
- ✅ 健康检查接口正常响应
- ✅ 所有配置更新完成

迁移时间：2025-07-20 15:25:00 