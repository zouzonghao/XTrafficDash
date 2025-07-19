-- X-UI 流量数据数据库表结构
-- 创建时间: 2024-01-01
-- 描述: 存储X-UI服务的流量数据，包括入站流量和客户端流量

-- 1. 服务表 - 记录每个IP对应的X-UI服务
CREATE TABLE IF NOT EXISTS services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip_address TEXT NOT NULL UNIQUE,           -- 服务IP地址
    service_name TEXT,                         -- 服务名称（可选）
    first_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- 首次发现时间
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,   -- 最后活跃时间
    status TEXT DEFAULT 'active',              -- 服务状态：active, inactive
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. 入站流量表 - 记录每个入站端口的流量数据
CREATE TABLE IF NOT EXISTS inbound_traffics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service_id INTEGER NOT NULL,               -- 关联服务ID
    tag TEXT NOT NULL,                         -- 端口标识（如：inbound-39062）
    port INTEGER,                              -- 端口号（从tag中解析）
    up BIGINT DEFAULT 0,                       -- 今日上传流量（字节）
    down BIGINT DEFAULT 0,                     -- 今日下载流量（字节）
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- 最后更新时间
    status TEXT DEFAULT 'active',              -- 状态：active, inactive
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
    UNIQUE(service_id, tag)
);

-- 3. 客户端流量表 - 记录每个用户的流量数据
CREATE TABLE IF NOT EXISTS client_traffics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service_id INTEGER NOT NULL,               -- 关联服务ID
    email TEXT NOT NULL,                       -- 用户邮箱
    up BIGINT DEFAULT 0,                       -- 今日上传流量（字节）
    down BIGINT DEFAULT 0,                     -- 今日下载流量（字节）
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- 最后更新时间
    status TEXT DEFAULT 'active',              -- 状态：active, inactive
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
    UNIQUE(service_id, email)
);

-- 4. 入站流量历史记录表 - 每日流量统计
CREATE TABLE IF NOT EXISTS inbound_traffic_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    inbound_traffic_id INTEGER NOT NULL,       -- 关联入站流量ID
    service_id INTEGER NOT NULL,               -- 关联服务ID
    tag TEXT NOT NULL,                         -- 端口标识
    date DATE NOT NULL,                        -- 日期（YYYY-MM-DD）
    daily_up BIGINT DEFAULT 0,                 -- 当日上传流量
    daily_down BIGINT DEFAULT 0,               -- 当日下载流量
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (inbound_traffic_id) REFERENCES inbound_traffics(id) ON DELETE CASCADE,
    FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
    UNIQUE(inbound_traffic_id, date)
);

-- 5. 原始请求记录表 - 记录所有接收到的原始数据
CREATE TABLE IF NOT EXISTS raw_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service_id INTEGER NOT NULL,               -- 关联服务ID
    client_ip TEXT NOT NULL,                   -- 客户端IP
    user_agent TEXT,                           -- User-Agent
    request_body TEXT,                         -- 原始请求体
    parsed_data TEXT,                          -- 解析后的JSON数据
    received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- 接收时间
    processed BOOLEAN DEFAULT FALSE,           -- 是否已处理
    FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_services_ip ON services(ip_address);
CREATE INDEX IF NOT EXISTS idx_services_status ON services(status);
CREATE INDEX IF NOT EXISTS idx_inbound_traffics_service_tag ON inbound_traffics(service_id, tag);
CREATE INDEX IF NOT EXISTS idx_inbound_traffics_status ON inbound_traffics(status);
CREATE INDEX IF NOT EXISTS idx_client_traffics_service_email ON client_traffics(service_id, email);
CREATE INDEX IF NOT EXISTS idx_client_traffics_status ON client_traffics(status);
CREATE INDEX IF NOT EXISTS idx_inbound_history_date ON inbound_traffic_history(date);
CREATE INDEX IF NOT EXISTS idx_inbound_history_service_date ON inbound_traffic_history(service_id, date);
CREATE INDEX IF NOT EXISTS idx_raw_requests_service_time ON raw_requests(service_id, received_at);

-- 创建触发器，自动更新updated_at字段
CREATE TRIGGER IF NOT EXISTS update_services_updated_at 
    AFTER UPDATE ON services
    FOR EACH ROW
BEGIN
    UPDATE services SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_inbound_traffics_updated_at 
    AFTER UPDATE ON inbound_traffics
    FOR EACH ROW
BEGIN
    UPDATE inbound_traffics SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_client_traffics_updated_at 
    AFTER UPDATE ON client_traffics
    FOR EACH ROW
BEGIN
    UPDATE client_traffics SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- 创建视图，方便查询汇总数据
CREATE VIEW IF NOT EXISTS v_service_summary AS
SELECT 
    s.id,
    s.ip_address,
    s.service_name,
    s.last_seen,
    COUNT(DISTINCT it.id) as inbound_count,
    COUNT(DISTINCT ct.id) as client_count,
    SUM(it.up) as total_inbound_up,
    SUM(it.down) as total_inbound_down,
    SUM(ct.up) as total_client_up,
    SUM(ct.down) as total_client_down
FROM services s
LEFT JOIN inbound_traffics it ON s.id = it.service_id AND it.status = 'active'
LEFT JOIN client_traffics ct ON s.id = ct.service_id AND ct.status = 'active'
GROUP BY s.id, s.ip_address, s.service_name, s.last_seen;

-- 创建视图，显示今日流量统计
CREATE VIEW IF NOT EXISTS v_today_traffic AS
SELECT 
    s.ip_address,
    it.tag,
    it.up as current_up,
    it.down as current_down,
    ith.daily_up as today_up,
    ith.daily_down as today_down,
    it.last_updated
FROM services s
JOIN inbound_traffics it ON s.id = it.service_id
LEFT JOIN inbound_traffic_history ith ON it.id = ith.inbound_traffic_id 
    AND ith.date = DATE('now')
WHERE it.status = 'active'
ORDER BY s.ip_address, it.tag; 