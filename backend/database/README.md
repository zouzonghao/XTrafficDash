# X-UI 流量数据库设计文档

## 概述

本数据库用于存储X-UI面板的流量数据，包括服务信息、入站流量、客户端流量和历史记录。

## 数据库表结构

### 1. services（服务表）

记录每个IP对应的X-UI服务信息。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | INTEGER | 主键，自增 |
| ip_address | TEXT | 服务IP地址（唯一） |
| service_name | TEXT | 服务名称 |
| first_seen | TIMESTAMP | 首次发现时间 |
| last_seen | TIMESTAMP | 最后活跃时间 |
| status | TEXT | 服务状态（active/inactive） |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### 2. inbound_traffics（入站流量表）

记录每个入站端口的流量数据。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | INTEGER | 主键，自增 |
| service_id | INTEGER | 关联服务ID |
| tag | TEXT | 端口标识（如：inbound-39062） |
| port | INTEGER | 端口号（从tag中解析） |
| up | BIGINT | 累计上传流量（字节） |
| down | BIGINT | 累计下载流量（字节） |
| up_temp | BIGINT | 临时上传流量（当前周期） |
| down_temp | BIGINT | 临时下载流量（当前周期） |
| last_updated | TIMESTAMP | 最后更新时间 |
| status | TEXT | 状态（active/inactive） |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### 3. client_traffics（客户端流量表）

记录每个用户的流量数据。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | INTEGER | 主键，自增 |
| service_id | INTEGER | 关联服务ID |
| email | TEXT | 用户邮箱 |
| up | BIGINT | 累计上传流量（字节） |
| down | BIGINT | 累计下载流量（字节） |
| up_temp | BIGINT | 临时上传流量（当前周期） |
| down_temp | BIGINT | 临时下载流量（当前周期） |
| last_updated | TIMESTAMP | 最后更新时间 |
| status | TEXT | 状态（active/inactive） |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### 4. inbound_traffic_history（入站流量历史记录表）

每日流量统计记录。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | INTEGER | 主键，自增 |
| inbound_traffic_id | INTEGER | 关联入站流量ID |
| service_id | INTEGER | 关联服务ID |
| tag | TEXT | 端口标识 |
| date | DATE | 日期（YYYY-MM-DD） |
| daily_up | BIGINT | 当日上传流量 |
| daily_down | BIGINT | 当日下载流量 |
| created_at | TIMESTAMP | 创建时间 |

### 5. raw_requests（原始请求记录表）

记录所有接收到的原始数据。

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | INTEGER | 主键，自增 |
| service_id | INTEGER | 关联服务ID |
| client_ip | TEXT | 客户端IP |
| user_agent | TEXT | User-Agent |
| request_body | TEXT | 原始请求体 |
| parsed_data | TEXT | 解析后的JSON数据 |
| received_at | TIMESTAMP | 接收时间 |
| processed | BOOLEAN | 是否已处理 |

## 流量处理逻辑

### 入站流量处理规则

1. **流量不为0时**：将流量值存储到 `up_temp` 和 `down_temp` 字段
2. **流量为0时**：将 `up_temp` 和 `down_temp` 的值加到 `up` 和 `down` 字段，然后清零临时字段

### 客户端流量处理规则

与入站流量处理规则相同。

### 每日统计

每日0点自动执行统计任务：
1. 将所有 `up_temp` 和 `down_temp` 的值加到对应的 `up` 和 `down` 字段
2. 将累计流量记录到历史表中
3. 清零临时字段

## API接口

### 服务管理

#### 获取所有服务列表
```
GET /api/db/services
```

响应示例：
```json
{
  "success": true,
  "message": "获取服务列表成功",
  "data": [
    {
      "id": 1,
      "ip_address": "192.74.226.78",
      "service_name": "X-UI-Service-192.74.226.78",
      "last_seen": "2025-07-19T10:26:34Z",
      "inbound_count": 3,
      "client_count": 2,
      "total_inbound_up": 22469,
      "total_inbound_down": 28463,
      "total_client_up": 17880,
      "total_client_down": 27803
    }
  ]
}
```

#### 获取服务详情
```
GET /api/db/services/:id
```

#### 获取服务流量详情
```
GET /api/db/services/:id/traffic
```

### 流量统计

#### 获取流量汇总
```
GET /api/db/traffic/summary
```

响应示例：
```json
{
  "success": true,
  "message": "获取流量汇总成功",
  "data": {
    "total_services": 2,
    "total_up": 44938,
    "total_down": 56926,
    "total_traffic": 101864,
    "services": [...]
  }
}
```

#### 获取流量历史记录
```
GET /api/db/traffic/history?service_id=1&tag=inbound-39062&start_date=2024-01-01&end_date=2024-01-31
```

### 原始数据

#### 获取原始请求记录
```
GET /api/db/raw-requests?service_id=1&limit=50
```

#### 获取原始请求详情
```
GET /api/db/raw-requests/:id
```

### 系统管理

#### 手动触发每日统计
```
POST /api/db/daily-summary
```

## 查询示例

### 查看所有活跃服务
```sql
SELECT * FROM services WHERE status = 'active' ORDER BY last_seen DESC;
```

### 查看某个服务的所有入站流量
```sql
SELECT * FROM inbound_traffics 
WHERE service_id = 1 AND status = 'active' 
ORDER BY tag;
```

### 查看某个端口的流量历史
```sql
SELECT * FROM inbound_traffic_history 
WHERE tag = 'inbound-39062' 
ORDER BY date DESC;
```

### 查看今日流量统计
```sql
SELECT 
    s.ip_address,
    it.tag,
    it.up + it.up_temp as current_up,
    it.down + it.down_temp as current_down
FROM services s
JOIN inbound_traffics it ON s.id = it.service_id
WHERE it.status = 'active'
ORDER BY s.ip_address, it.tag;
```

## 索引设计

为提高查询性能，创建了以下索引：

- `idx_services_ip`: 服务IP地址索引
- `idx_inbound_traffics_service_tag`: 入站流量服务+标签索引
- `idx_client_traffics_service_email`: 客户端流量服务+邮箱索引
- `idx_inbound_history_date`: 历史记录日期索引
- `idx_raw_requests_service_time`: 原始请求服务+时间索引

## 触发器

自动更新 `updated_at` 字段的触发器：

- `update_services_updated_at`
- `update_inbound_traffics_updated_at`
- `update_client_traffics_updated_at`

## 视图

### v_service_summary
服务汇总视图，包含服务基本信息和流量统计。

### v_today_traffic
今日流量视图，显示当前流量和历史记录。

## 数据维护

### 定期清理
建议定期清理以下数据：
- 超过30天的原始请求记录
- 超过90天的历史流量记录
- 超过7天未活跃的服务

### 备份策略
- 每日自动备份数据库文件
- 保留最近30天的备份
- 重要数据变更前手动备份 