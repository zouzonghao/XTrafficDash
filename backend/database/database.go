package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// 数据库结构体
type Database struct {
	db *sql.DB
}

// 流量数据结构体
type TrafficData struct {
	ClientTraffics  []ClientTraffic  `json:"clientTraffics"`
	InboundTraffics []InboundTraffic `json:"inboundTraffics"`
}

// 客户端流量结构体
type ClientTraffic struct {
	ID         int    `json:"id"`
	InboundID  int    `json:"inboundId"`
	Enable     bool   `json:"enable"`
	Email      string `json:"email"`
	Up         int64  `json:"up"`
	Down       int64  `json:"down"`
	ExpiryTime int64  `json:"expiryTime"`
	Total      int64  `json:"total"`
	Reset      int64  `json:"reset"`
}

// 入站流量结构体
type InboundTraffic struct {
	IsInbound  bool   `json:"IsInbound"`
	IsOutbound bool   `json:"IsOutbound"`
	Tag        string `json:"Tag"`
	Up         int64  `json:"Up"`
	Down       int64  `json:"Down"`
}

// 服务信息结构体
type Service struct {
	ID          int       `json:"id"`
	IPAddress   string    `json:"ip_address"`
	ServiceName string    `json:"service_name"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	Status      string    `json:"status"`
}

// 入站流量记录结构体
type InboundTrafficRecord struct {
	ID          int       `json:"id"`
	ServiceID   int       `json:"service_id"`
	Tag         string    `json:"tag"`
	Port        int       `json:"port"`
	CustomName  string    `json:"custom_name"`
	Up          int64     `json:"up"`
	Down        int64     `json:"down"`
	LastUpdated time.Time `json:"last_updated"`
	Status      string    `json:"status"`
}

// 客户端流量记录结构体
type ClientTrafficRecord struct {
	ID          int       `json:"id"`
	ServiceID   int       `json:"service_id"`
	Email       string    `json:"email"`
	CustomName  string    `json:"custom_name"`
	Up          int64     `json:"up"`
	Down        int64     `json:"down"`
	LastUpdated time.Time `json:"last_updated"`
	Status      string    `json:"status"`
}

// 打开数据库连接
func OpenDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	// 设置时区为本地时间
	if _, err := db.Exec("PRAGMA timezone = 'local'"); err != nil {
		return nil, fmt.Errorf("设置时区失败: %v", err)
	}

	// 初始化数据库表
	if err := initDatabase(db); err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %v", err)
	}

	// 执行数据库迁移
	if err := migrateDatabase(db); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %v", err)
	}

	return &Database{db: db}, nil
}

// 数据库迁移函数
func migrateDatabase(db *sql.DB) error {
	// 检查并添加 services 表的 custom_name 字段
	if !columnExists(db, "services", "custom_name") {
		_, err := db.Exec("ALTER TABLE services ADD COLUMN custom_name TEXT")
		if err != nil {
			return fmt.Errorf("添加 services.custom_name 字段失败: %v", err)
		}
	}

	// 检查并添加 inbound_traffics 表的 custom_name 字段
	if !columnExists(db, "inbound_traffics", "custom_name") {
		_, err := db.Exec("ALTER TABLE inbound_traffics ADD COLUMN custom_name TEXT")
		if err != nil {
			return fmt.Errorf("添加 inbound_traffics.custom_name 字段失败: %v", err)
		}
	}

	// 检查并添加 client_traffics 表的 custom_name 字段
	if !columnExists(db, "client_traffics", "custom_name") {
		_, err := db.Exec("ALTER TABLE client_traffics ADD COLUMN custom_name TEXT")
		if err != nil {
			return fmt.Errorf("添加 client_traffics.custom_name 字段失败: %v", err)
		}
	}

	return nil
}

// 检查列是否存在
func columnExists(db *sql.DB, tableName, columnName string) bool {
	query := `
		SELECT COUNT(*) 
		FROM pragma_table_info(?) 
		WHERE name = ?
	`
	var count int
	err := db.QueryRow(query, tableName, columnName).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

// 关闭数据库连接
func (d *Database) Close() error {
	return d.db.Close()
}

// 初始化数据库表
func initDatabase(db *sql.DB) error {
	// 读取SQL文件内容
	schemaSQL := `
	-- X-UI 流量数据数据库表结构
	-- 创建时间: 2024-01-01
	-- 描述: 存储X-UI服务的流量数据，包括入站流量和客户端流量

	-- 1. 服务表 - 记录每个IP对应的X-UI服务
	CREATE TABLE IF NOT EXISTS services (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip_address TEXT NOT NULL UNIQUE,
		service_name TEXT,
		custom_name TEXT,
		first_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		status TEXT DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- 2. 入站流量表 - 记录每个入站端口的流量数据
	CREATE TABLE IF NOT EXISTS inbound_traffics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service_id INTEGER NOT NULL,
		tag TEXT NOT NULL,
		port INTEGER,
		custom_name TEXT,
		up BIGINT DEFAULT 0,
		down BIGINT DEFAULT 0,
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		status TEXT DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
		UNIQUE(service_id, tag)
	);

	-- 3. 客户端流量表 - 记录每个用户的流量数据
	CREATE TABLE IF NOT EXISTS client_traffics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service_id INTEGER NOT NULL,
		email TEXT NOT NULL,
		custom_name TEXT,
		up BIGINT DEFAULT 0,
		down BIGINT DEFAULT 0,
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		status TEXT DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
		UNIQUE(service_id, email)
	);

	-- 4. 入站流量历史记录表 - 每日流量统计
	CREATE TABLE IF NOT EXISTS inbound_traffic_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		inbound_traffic_id INTEGER NOT NULL,
		service_id INTEGER NOT NULL,
		tag TEXT NOT NULL,
		date DATE NOT NULL,
		daily_up BIGINT DEFAULT 0,
		daily_down BIGINT DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (inbound_traffic_id) REFERENCES inbound_traffics(id) ON DELETE CASCADE,
		FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
		UNIQUE(inbound_traffic_id, date)
	);

	-- 5. 客户端流量历史记录表 - 每日流量统计
	CREATE TABLE IF NOT EXISTS client_traffic_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client_traffic_id INTEGER NOT NULL,
		service_id INTEGER NOT NULL,
		email TEXT NOT NULL,
		date DATE NOT NULL,
		daily_up BIGINT DEFAULT 0,
		daily_down BIGINT DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (client_traffic_id) REFERENCES client_traffics(id) ON DELETE CASCADE,
		FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE,
		UNIQUE(client_traffic_id, date)
	);

	-- 6. 原始请求记录表 - 记录所有接收到的原始数据
	CREATE TABLE IF NOT EXISTS raw_requests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service_id INTEGER NOT NULL,
		client_ip TEXT NOT NULL,
		user_agent TEXT,
		request_body TEXT,
		parsed_data TEXT,
		received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		processed BOOLEAN DEFAULT FALSE,
		FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE
	);

	-- 创建索引
	CREATE INDEX IF NOT EXISTS idx_services_ip ON services(ip_address);
	CREATE INDEX IF NOT EXISTS idx_inbound_traffics_service_tag ON inbound_traffics(service_id, tag);
	CREATE INDEX IF NOT EXISTS idx_client_traffics_service_email ON client_traffics(service_id, email);
	CREATE INDEX IF NOT EXISTS idx_inbound_history_date ON inbound_traffic_history(date);
	CREATE INDEX IF NOT EXISTS idx_client_history_date ON client_traffic_history(date);
	CREATE INDEX IF NOT EXISTS idx_raw_requests_service_time ON raw_requests(service_id, received_at);
	`

	// 执行SQL语句
	_, err := db.Exec(schemaSQL)
	return err
}

// 处理流量数据
func (d *Database) ProcessTrafficData(clientIP string, userAgent string, requestBody string, trafficData *TrafficData) error {
	// 开始事务
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer tx.Rollback()

	// 1. 获取或创建服务记录
	serviceID, err := d.getOrCreateService(tx, clientIP)
	if err != nil {
		return fmt.Errorf("获取或创建服务失败: %v", err)
	}

	// 2. 记录原始请求
	parsedData, _ := json.Marshal(trafficData)
	err = d.recordRawRequest(tx, serviceID, clientIP, userAgent, requestBody, string(parsedData))
	if err != nil {
		return fmt.Errorf("记录原始请求失败: %v", err)
	}

	// 3. 处理入站流量数据并记录有流量的端口
	err = d.processInboundTraffics(tx, serviceID, trafficData.InboundTraffics)
	if err != nil {
		return fmt.Errorf("处理入站流量失败: %v", err)
	}

	// 4. 处理客户端流量数据
	err = d.processClientTraffics(tx, serviceID, trafficData.ClientTraffics)
	if err != nil {
		return fmt.Errorf("处理客户端流量失败: %v", err)
	}

	// 5. 只要有数据包发来就更新节点最后活跃时间（包括心跳数据）
	err = d.updateServiceLastSeen(tx, serviceID)
	if err != nil {
		return fmt.Errorf("更新服务最后活跃时间失败: %v", err)
	}

	// 提交事务
	return tx.Commit()
}

// 获取或创建服务记录
func (d *Database) getOrCreateService(tx *sql.Tx, ipAddress string) (int, error) {
	var serviceID int

	// 先尝试查找现有服务
	err := tx.QueryRow("SELECT id FROM services WHERE ip_address = ?", ipAddress).Scan(&serviceID)
	if err == sql.ErrNoRows {
		// 创建新服务
		result, err := tx.Exec(`
			INSERT INTO services (ip_address, service_name, first_seen, last_seen, status)
			VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'active')
		`, ipAddress, fmt.Sprintf("X-UI-Service-%s", ipAddress))
		if err != nil {
			return 0, err
		}

		serviceID64, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		serviceID = int(serviceID64)
	} else if err != nil {
		return 0, err
	}

	return serviceID, nil
}

// 记录原始请求
func (d *Database) recordRawRequest(tx *sql.Tx, serviceID int, clientIP, userAgent, requestBody, parsedData string) error {
	_, err := tx.Exec(`
		INSERT INTO raw_requests (service_id, client_ip, user_agent, request_body, parsed_data, received_at, processed)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, TRUE)
	`, serviceID, clientIP, userAgent, requestBody, parsedData)
	return err
}

// 处理入站流量数据
func (d *Database) processInboundTraffics(tx *sql.Tx, serviceID int, inboundTraffics []InboundTraffic) error {
	// 记录有流量的端口信息
	var activePorts []string

	for _, traffic := range inboundTraffics {
		// 只处理入站流量
		if !traffic.IsInbound {
			continue
		}

		// 解析端口号
		port := d.extractPortFromTag(traffic.Tag)

		// 如果有流量，记录端口信息
		if traffic.Up > 0 || traffic.Down > 0 {
			activePorts = append(activePorts, fmt.Sprintf("端口%d(上传:%s,下载:%s)",
				port,
				d.formatBytes(traffic.Up),
				d.formatBytes(traffic.Down)))
		}

		// 获取或创建入站流量记录
		var recordID int
		err := tx.QueryRow(`
			SELECT id FROM inbound_traffics 
			WHERE service_id = ? AND tag = ?
		`, serviceID, traffic.Tag).Scan(&recordID)

		if err == sql.ErrNoRows {
			// 创建新记录
			result, err := tx.Exec(`
				INSERT INTO inbound_traffics (service_id, tag, port, up, down, last_updated, status)
				VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, 'active')
			`, serviceID, traffic.Tag, port, traffic.Up, traffic.Down)
			if err != nil {
				return err
			}
			recordID64, _ := result.LastInsertId()
			recordID = int(recordID64)
		} else if err != nil {
			return err
		} else {
			// 更新现有记录
			err = d.updateInboundTraffic(tx, recordID, traffic.Up, traffic.Down)
			if err != nil {
				return err
			}
		}
	}

	// 如果有活跃端口，输出日志
	if len(activePorts) > 0 {
		fmt.Printf("有流量的端口: %s\n", strings.Join(activePorts, ", "))
	}

	return nil
}

// 处理客户端流量数据
func (d *Database) processClientTraffics(tx *sql.Tx, serviceID int, clientTraffics []ClientTraffic) error {
	for _, traffic := range clientTraffics {
		// 获取或创建客户端流量记录
		var recordID int
		err := tx.QueryRow(`
			SELECT id FROM client_traffics 
			WHERE service_id = ? AND email = ?
		`, serviceID, traffic.Email).Scan(&recordID)

		if err == sql.ErrNoRows {
			// 创建新记录
			result, err := tx.Exec(`
				INSERT INTO client_traffics (service_id, email, up, down, last_updated, status)
				VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, 'active')
			`, serviceID, traffic.Email, traffic.Up, traffic.Down)
			if err != nil {
				return err
			}
			recordID64, _ := result.LastInsertId()
			recordID = int(recordID64)
		} else if err != nil {
			return err
		} else {
			// 更新现有记录
			err = d.updateClientTraffic(tx, recordID, traffic.Up, traffic.Down)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// 更新入站流量记录
func (d *Database) updateInboundTraffic(tx *sql.Tx, recordID int, up, down int64) error {
	// 只有当流量不为0时才更新last_updated时间
	if up > 0 || down > 0 {
		_, err := tx.Exec(`
			UPDATE inbound_traffics 
			SET up = up + ?, down = down + ?, last_updated = CURRENT_TIMESTAMP
			WHERE id = ?
		`, up, down, recordID)
		return err
	} else {
		// 流量为0时，只更新流量数据，不更新时间
		_, err := tx.Exec(`
			UPDATE inbound_traffics 
			SET up = up + ?, down = down + ?
			WHERE id = ?
		`, up, down, recordID)
		return err
	}
}

// 更新客户端流量记录
func (d *Database) updateClientTraffic(tx *sql.Tx, recordID int, up, down int64) error {
	// 只有当流量不为0时才更新last_updated时间
	if up > 0 || down > 0 {
		_, err := tx.Exec(`
			UPDATE client_traffics 
			SET up = up + ?, down = down + ?, last_updated = CURRENT_TIMESTAMP
			WHERE id = ?
		`, up, down, recordID)
		return err
	} else {
		// 流量为0时，只更新流量数据，不更新时间
		_, err := tx.Exec(`
			UPDATE client_traffics 
			SET up = up + ?, down = down + ?
			WHERE id = ?
		`, up, down, recordID)
		return err
	}
}

// 更新服务最后活跃时间
func (d *Database) updateServiceLastSeen(tx *sql.Tx, serviceID int) error {
	_, err := tx.Exec(`
		UPDATE services 
		SET last_seen = CURRENT_TIMESTAMP 
		WHERE id = ?
	`, serviceID)
	return err
}

// 从tag中提取端口号
func (d *Database) extractPortFromTag(tag string) int {
	re := regexp.MustCompile(`inbound-(\d+)`)
	matches := re.FindStringSubmatch(tag)
	if len(matches) > 1 {
		if port, err := strconv.Atoi(matches[1]); err == nil {
			return port
		}
	}
	return 0
}

// IP地址脱敏处理
func (d *Database) maskIPAddress(ip string) string {
	// 处理IPv4地址
	if strings.Contains(ip, ".") {
		parts := strings.Split(ip, ".")
		if len(parts) == 4 {
			// 保留前两段，后两段用xxx替换
			return parts[0] + "." + parts[1] + ".xxx.xxx"
		}
	}

	// 处理IPv6地址
	if strings.Contains(ip, ":") {
		parts := strings.Split(ip, ":")
		if len(parts) >= 4 {
			// 保留前两段，其余用xxx替换
			return parts[0] + ":" + parts[1] + ":xxx:xxx"
		}
	}

	// 如果无法解析，返回原IP
	return ip
}

// 格式化字节数
func (d *Database) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// 获取服务汇总信息
func (d *Database) GetServiceSummary() ([]map[string]interface{}, error) {
	// 首先获取所有服务的基本信息
	serviceRows, err := d.db.Query(`
		SELECT 
			s.id,
			s.ip_address,
			s.service_name,
			s.custom_name,
			s.last_seen,
					CASE 
			WHEN (strftime('%s', 'now') - strftime('%s', s.last_seen)) <= 30 THEN 'active'
			ELSE 'inactive'
		END as status
		FROM services s
		ORDER BY s.last_seen DESC
	`)
	if err != nil {
		return nil, err
	}
	defer serviceRows.Close()

	var results []map[string]interface{}
	serviceMap := make(map[int]map[string]interface{})

	// 处理服务基本信息
	for serviceRows.Next() {
		var id int
		var ipAddress, serviceName, lastSeen, status string
		var customName sql.NullString

		scanErr := serviceRows.Scan(&id, &ipAddress, &serviceName, &customName, &lastSeen, &status)
		if scanErr != nil {
			return nil, scanErr
		}

		result := map[string]interface{}{
			"id":                 id,
			"ip_address":         d.maskIPAddress(ipAddress),
			"service_name":       serviceName,
			"custom_name":        customName.String,
			"last_seen":          lastSeen,
			"status":             status,
			"inbound_count":      0,
			"client_count":       0,
			"total_inbound_up":   0,
			"total_inbound_down": 0,
			"total_client_up":    0,
			"total_client_down":  0,
		}
		serviceMap[id] = result
		results = append(results, result)
	}

	// 查询入站流量数据
	inboundRows, err := d.db.Query(`
		SELECT 
			service_id,
			COUNT(*) as inbound_count,
			SUM(up) as total_inbound_up,
			SUM(down) as total_inbound_down
		FROM inbound_traffics 
		WHERE status = 'active'
		GROUP BY service_id
	`)
	if err != nil {
		return nil, err
	}
	defer inboundRows.Close()

	// 处理入站流量数据
	for inboundRows.Next() {
		var serviceID int
		var inboundCount int
		var totalInboundUp, totalInboundDown sql.NullInt64

		scanErr := inboundRows.Scan(&serviceID, &inboundCount, &totalInboundUp, &totalInboundDown)
		if scanErr != nil {
			return nil, scanErr
		}

		if service, exists := serviceMap[serviceID]; exists {
			service["inbound_count"] = inboundCount
			service["total_inbound_up"] = totalInboundUp.Int64
			service["total_inbound_down"] = totalInboundDown.Int64
		}
	}

	// 查询客户端流量数据
	clientRows, err := d.db.Query(`
		SELECT 
			service_id,
			COUNT(*) as client_count,
			SUM(up) as total_client_up,
			SUM(down) as total_client_down
		FROM client_traffics 
		WHERE status = 'active'
		GROUP BY service_id
	`)
	if err != nil {
		return nil, err
	}
	defer clientRows.Close()

	// 处理客户端流量数据
	for clientRows.Next() {
		var serviceID int
		var clientCount int
		var totalClientUp, totalClientDown sql.NullInt64

		scanErr := clientRows.Scan(&serviceID, &clientCount, &totalClientUp, &totalClientDown)
		if scanErr != nil {
			return nil, scanErr
		}

		if service, exists := serviceMap[serviceID]; exists {
			service["client_count"] = clientCount
			service["total_client_up"] = totalClientUp.Int64
			service["total_client_down"] = totalClientDown.Int64
		}
	}

	return results, nil
}

// 获取指定服务的详细流量信息
func (d *Database) GetServiceTraffic(serviceID int) (map[string]interface{}, error) {
	// 获取服务基本信息
	var service Service
	var rawIPAddress string
	var customName sql.NullString
	err := d.db.QueryRow(`
		SELECT id, ip_address, service_name, custom_name, first_seen, last_seen, status
		FROM services WHERE id = ?
	`, serviceID).Scan(&service.ID, &rawIPAddress, &service.ServiceName, &customName,
		&service.FirstSeen, &service.LastSeen, &service.Status)
	if err != nil {
		return nil, err
	}

	// 对IP地址进行脱敏
	service.IPAddress = d.maskIPAddress(rawIPAddress)

	// 获取入站流量
	inboundRows, err := d.db.Query(`
		SELECT id, service_id, tag, port, custom_name, up, down, last_updated, status
		FROM inbound_traffics WHERE service_id = ? AND status = 'active'
		ORDER BY tag
	`, serviceID)
	if err != nil {
		return nil, err
	}
	defer inboundRows.Close()

	var inboundTraffics []InboundTrafficRecord
	for inboundRows.Next() {
		var record InboundTrafficRecord
		var customName sql.NullString
		err := inboundRows.Scan(&record.ID, &record.ServiceID, &record.Tag, &record.Port, &customName,
			&record.Up, &record.Down, &record.LastUpdated, &record.Status)
		if err != nil {
			return nil, err
		}
		// 保存自定义名称
		record.CustomName = customName.String
		inboundTraffics = append(inboundTraffics, record)
	}

	// 获取客户端流量
	clientRows, err := d.db.Query(`
		SELECT id, service_id, email, custom_name, up, down, last_updated, status
		FROM client_traffics WHERE service_id = ? AND status = 'active'
		ORDER BY email
	`, serviceID)
	if err != nil {
		return nil, err
	}
	defer clientRows.Close()

	var clientTraffics []ClientTrafficRecord
	for clientRows.Next() {
		var record ClientTrafficRecord
		var customName sql.NullString
		err := clientRows.Scan(&record.ID, &record.ServiceID, &record.Email, &customName, &record.Up,
			&record.Down, &record.LastUpdated, &record.Status)
		if err != nil {
			return nil, err
		}
		// 保存自定义名称
		record.CustomName = customName.String
		clientTraffics = append(clientTraffics, record)
	}

	result := map[string]interface{}{
		"service":          service,
		"inbound_traffics": inboundTraffics,
		"client_traffics":  clientTraffics,
	}

	return result, nil
}

// 删除服务及其所有相关数据
func (d *Database) DeleteService(serviceID int) error {
	log.Printf("开始删除服务ID: %d", serviceID)

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 删除历史记录
	_, err = tx.Exec("DELETE FROM inbound_traffic_history WHERE service_id = ?", serviceID)
	if err != nil {
		return fmt.Errorf("删除历史记录失败: %v", err)
	}

	// 删除入站流量记录
	_, err = tx.Exec("DELETE FROM inbound_traffics WHERE service_id = ?", serviceID)
	if err != nil {
		return fmt.Errorf("删除入站流量记录失败: %v", err)
	}

	// 删除客户端流量记录
	_, err = tx.Exec("DELETE FROM client_traffics WHERE service_id = ?", serviceID)
	if err != nil {
		return fmt.Errorf("删除客户端流量记录失败: %v", err)
	}

	// 删除原始请求记录
	_, err = tx.Exec("DELETE FROM raw_requests WHERE service_id = ?", serviceID)
	if err != nil {
		return fmt.Errorf("删除原始请求记录失败: %v", err)
	}

	// 删除服务记录
	_, err = tx.Exec("DELETE FROM services WHERE id = ?", serviceID)
	if err != nil {
		return fmt.Errorf("删除服务记录失败: %v", err)
	}

	log.Printf("服务ID %d 删除成功", serviceID)
	return tx.Commit()
}

// 每日流量统计任务（需要在每日0点执行）
func (d *Database) DailyTrafficSummary() error {
	log.Println("开始执行每日流量统计...")

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 处理入站流量记录
	err = d.processDailyInboundTraffic(tx)
	if err != nil {
		return err
	}

	// 处理客户端流量记录
	err = d.processDailyClientTraffic(tx)
	if err != nil {
		return err
	}

	log.Println("每日流量统计完成")
	return tx.Commit()
}

// 处理入站流量的每日统计
func (d *Database) processDailyInboundTraffic(tx *sql.Tx) error {
	// 获取所有活跃的入站流量记录
	rows, err := tx.Query(`
		SELECT id, service_id, tag, up, down
		FROM inbound_traffics 
		WHERE status = 'active'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, serviceID int
		var tag string
		var dailyUp, dailyDown int64

		err := rows.Scan(&id, &serviceID, &tag, &dailyUp, &dailyDown)
		if err != nil {
			return err
		}

		// 如果今日有流量，记录到历史表
		if dailyUp > 0 || dailyDown > 0 {
			_, err = tx.Exec(`
				INSERT OR REPLACE INTO inbound_traffic_history 
				(inbound_traffic_id, service_id, tag, date, daily_up, daily_down, created_at)
				VALUES (?, ?, ?, DATE('now', '-1 day'), ?, ?, CURRENT_TIMESTAMP)
			`, id, serviceID, tag, dailyUp, dailyDown)
			if err != nil {
				return err
			}

			log.Printf("端口 %s: 记录昨日流量 up=%d, down=%d", tag, dailyUp, dailyDown)
		}

		// 清零今日流量
		_, err = tx.Exec(`
			UPDATE inbound_traffics 
			SET up = 0, down = 0
			WHERE id = ?
		`, id)
		if err != nil {
			return err
		}

		log.Printf("端口 %s: 清零今日流量", tag)
	}

	return nil
}

// 处理客户端流量的每日统计
func (d *Database) processDailyClientTraffic(tx *sql.Tx) error {
	// 获取所有活跃的客户端流量记录
	rows, err := tx.Query(`
		SELECT id, service_id, email, up, down
		FROM client_traffics 
		WHERE status = 'active'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, serviceID int
		var email string
		var dailyUp, dailyDown int64

		err := rows.Scan(&id, &serviceID, &email, &dailyUp, &dailyDown)
		if err != nil {
			return err
		}

		// 如果今日有流量，记录到历史表
		if dailyUp > 0 || dailyDown > 0 {
			_, err = tx.Exec(`
				INSERT OR REPLACE INTO client_traffic_history 
				(client_traffic_id, service_id, email, date, daily_up, daily_down, created_at)
				VALUES (?, ?, ?, DATE('now', '-1 day'), ?, ?, CURRENT_TIMESTAMP)
			`, id, serviceID, email, dailyUp, dailyDown)
			if err != nil {
				return err
			}

			log.Printf("用户 %s: 记录昨日流量 up=%d, down=%d", email, dailyUp, dailyDown)
		}

		// 清零今日流量
		_, err = tx.Exec(`
			UPDATE client_traffics 
			SET up = 0, down = 0
			WHERE id = ?
		`, id)
		if err != nil {
			return err
		}

		log.Printf("用户 %s: 清零今日流量", email)
	}

	return nil
}
