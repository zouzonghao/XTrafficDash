package database

import (
	"database/sql"
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
	ID        int       `json:"id"`
	IPAddress string    `json:"ip_address"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	Status    string    `json:"status"`
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

// HY2配置结构体
// 用于存储hy2主动流量同步的参数

type Hy2Config struct {
	ID                int    `json:"id"`
	SourceAPIPassword string `json:"source_api_password"`
	SourceAPIHost     string `json:"source_api_host"`
	SourceAPIPort     string `json:"source_api_port"`
	TargetAPIURL      string `json:"target_api_url"`
}

// 打开数据库连接
func OpenDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %v", err)
	}

	// 配置数据库连接池
	db.SetMaxOpenConns(25)                 // 最大连接数
	db.SetMaxIdleConns(10)                 // 最大空闲连接数
	db.SetConnMaxLifetime(5 * time.Minute) // 连接最大生命周期
	db.SetConnMaxIdleTime(3 * time.Minute) // 空闲连接最大生命周期

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %v", err)
	}

	// 设置时区为本地时间
	if _, err := db.Exec("PRAGMA timezone = 'local'"); err != nil {
		return nil, fmt.Errorf("设置时区失败: %v", err)
	}

	// 优化SQLite性能
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("设置WAL模式失败: %v", err)
	}
	if _, err := db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
		return nil, fmt.Errorf("设置同步模式失败: %v", err)
	}
	if _, err := db.Exec("PRAGMA cache_size = 10000"); err != nil {
		return nil, fmt.Errorf("设置缓存大小失败: %v", err)
	}
	if _, err := db.Exec("PRAGMA temp_store = MEMORY"); err != nil {
		return nil, fmt.Errorf("设置临时存储失败: %v", err)
	}

	// 初始化数据库表
	if err := initDatabase(db); err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %v", err)
	}

	return &Database{db: db}, nil
}

// 关闭数据库连接
func (d *Database) Close() error {
	return d.db.Close()
}

// 初始化数据库表
func initDatabase(db *sql.DB) error {
	// 读取SQL文件内容
	schemaSQL := `
	-- XTrafficDash 流量数据数据库表结构
	-- 创建时间: 2024-01-01
	-- 描述: 存储X-UI服务的流量数据，包括入站流量和客户端流量

	-- 1. 服务表 - 记录每个IP对应的X-UI服务
	CREATE TABLE IF NOT EXISTS services (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip_address TEXT NOT NULL UNIQUE,
		custom_name TEXT,
		first_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		status TEXT DEFAULT 'active'
	);

	-- 2. 入站流量表 - 记录每个入站端口的流量数据
	CREATE TABLE IF NOT EXISTS inbound_traffics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service_id INTEGER NOT NULL,
		tag TEXT NOT NULL,
		port INTEGER,
		custom_name TEXT,
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		status TEXT DEFAULT 'active'
	);

	-- 3. 客户端流量表 - 记录每个用户的流量数据
	CREATE TABLE IF NOT EXISTS client_traffics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service_id INTEGER NOT NULL,
		email TEXT NOT NULL,
		custom_name TEXT,
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		status TEXT DEFAULT 'active'
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

	-- 6. HY2配置表
	CREATE TABLE IF NOT EXISTS hy2_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_api_password TEXT NOT NULL DEFAULT '',
		source_api_host TEXT NOT NULL DEFAULT '',
		source_api_port TEXT NOT NULL DEFAULT '',
		target_api_url TEXT NOT NULL DEFAULT ''
	);


	-- 创建索引
	CREATE INDEX IF NOT EXISTS idx_services_ip ON services(ip_address);
	CREATE INDEX IF NOT EXISTS idx_inbound_traffics_service_tag ON inbound_traffics(service_id, tag);
	CREATE INDEX IF NOT EXISTS idx_client_traffics_service_email ON client_traffics(service_id, email);
	CREATE INDEX IF NOT EXISTS idx_inbound_history_date ON inbound_traffic_history(date);
	CREATE INDEX IF NOT EXISTS idx_client_history_date ON client_traffic_history(date);
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

	// 2. 处理入站流量数据并记录有流量的端口
	err = d.processInboundTraffics(tx, serviceID, trafficData.InboundTraffics)
	if err != nil {
		return fmt.Errorf("处理入站流量失败: %v", err)
	}

	// 3. 处理客户端流量数据
	err = d.processClientTraffics(tx, serviceID, trafficData.ClientTraffics)
	if err != nil {
		return fmt.Errorf("处理客户端流量失败: %v", err)
	}

	// 4. 只要有数据包发来就更新节点最后活跃时间（包括心跳数据）
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
		now := time.Now()
		result, err := tx.Exec(`
			INSERT INTO services (ip_address, custom_name, first_seen, last_seen, status)
			VALUES (?, ?, ?, ?, 'active')
		`, ipAddress, "", now, now)
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

// 处理入站流量数据
func (d *Database) processInboundTraffics(tx *sql.Tx, serviceID int, inboundTraffics []InboundTraffic) error {
	var activePorts []string
	for _, traffic := range inboundTraffics {
		if !traffic.IsInbound {
			continue
		}
		port := d.extractPortFromTag(traffic.Tag)
		if traffic.Up > 0 || traffic.Down > 0 {
			activePorts = append(activePorts, fmt.Sprintf("端口%d(上传:%s,下载:%s)", port, d.formatBytes(traffic.Up), d.formatBytes(traffic.Down)))
		}
		// 获取或创建入站流量记录
		var recordID int
		err := tx.QueryRow(`SELECT id FROM inbound_traffics WHERE service_id = ? AND tag = ?`, serviceID, traffic.Tag).Scan(&recordID)
		if err == sql.ErrNoRows {
			now := time.Now()
			result, err := tx.Exec(`INSERT INTO inbound_traffics (service_id, tag, port, last_updated, status) VALUES (?, ?, ?, ?, 'active')`, serviceID, traffic.Tag, port, now)
			if err != nil {
				return err
			}
			recordID64, _ := result.LastInsertId()
			recordID = int(recordID64)
		} else if err != nil {
			return err
		}
		// upsert 到历史表，写入 date 用 localtime
		if traffic.Up > 0 || traffic.Down > 0 {
			_, err := tx.Exec(`
				INSERT INTO inbound_traffic_history (inbound_traffic_id, service_id, tag, date, daily_up, daily_down, created_at)
				VALUES (?, ?, ?, DATE('now', 'localtime'), ?, ?, ?)
				ON CONFLICT(inbound_traffic_id, date) DO UPDATE SET
					daily_up = daily_up + excluded.daily_up,
					daily_down = daily_down + excluded.daily_down
			`, recordID, serviceID, traffic.Tag, traffic.Up, traffic.Down, time.Now())
			if err != nil {
				return err
			}
			// 新增：有流量时更新 last_updated
			_, err = tx.Exec(`UPDATE inbound_traffics SET last_updated = ? WHERE id = ?`, time.Now(), recordID)
			if err != nil {
				return err
			}
		}
	}
	if len(activePorts) > 0 {
		fmt.Printf("活跃端口: %s\n", strings.Join(activePorts, ", "))
	}
	return nil
}

// 处理客户端流量数据
func (d *Database) processClientTraffics(tx *sql.Tx, serviceID int, clientTraffics []ClientTraffic) error {
	for _, traffic := range clientTraffics {
		var recordID int
		err := tx.QueryRow(`SELECT id FROM client_traffics WHERE service_id = ? AND email = ?`, serviceID, traffic.Email).Scan(&recordID)
		if err == sql.ErrNoRows {
			now := time.Now()
			result, err := tx.Exec(`INSERT INTO client_traffics (service_id, email, last_updated, status) VALUES (?, ?, ?, 'active')`, serviceID, traffic.Email, now)
			if err != nil {
				return err
			}
			recordID64, _ := result.LastInsertId()
			recordID = int(recordID64)
		} else if err != nil {
			return err
		}
		// upsert 到历史表，写入 date 用 localtime
		if traffic.Up > 0 || traffic.Down > 0 {
			_, err := tx.Exec(`
				INSERT INTO client_traffic_history (client_traffic_id, service_id, email, date, daily_up, daily_down, created_at)
				VALUES (?, ?, ?, DATE('now', 'localtime'), ?, ?, ?)
				ON CONFLICT(client_traffic_id, date) DO UPDATE SET
					daily_up = daily_up + excluded.daily_up,
					daily_down = daily_down + excluded.daily_down
			`, recordID, serviceID, traffic.Email, traffic.Up, traffic.Down, time.Now())
			if err != nil {
				return err
			}
			// 新增：有流量时更新 last_updated
			_, err = tx.Exec(`UPDATE client_traffics SET last_updated = ? WHERE id = ?`, time.Now(), recordID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// 更新服务最后活跃时间
func (d *Database) updateServiceLastSeen(tx *sql.Tx, serviceID int) error {
	now := time.Now()
	_, err := tx.Exec(`
		UPDATE services 
		SET last_seen = ? 
		WHERE id = ?
	`, now, serviceID)
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
	// 一次性查询所有统计信息，避免N+1问题
	rows, err := d.db.Query(`
		SELECT
			s.id,
			s.ip_address,
			s.custom_name,
			s.last_seen,
			CASE 
				WHEN (strftime('%s', 'now') - strftime('%s', s.last_seen)) <= 30 THEN 'active'
				ELSE 'inactive'
			END as status,
			COALESCE(it_counts.inbound_count, 0) as inbound_count,
			COALESCE(ct_counts.client_count, 0) as client_count,
			COALESCE(today_traffic.today_up, 0) as today_inbound_up,
			COALESCE(today_traffic.today_down, 0) as today_inbound_down
		FROM
			services s
		LEFT JOIN (
			SELECT service_id, COUNT(id) as inbound_count FROM inbound_traffics WHERE status = 'active' GROUP BY service_id
		) it_counts ON s.id = it_counts.service_id
		LEFT JOIN (
			SELECT service_id, COUNT(id) as client_count FROM client_traffics WHERE status = 'active' GROUP BY service_id
		) ct_counts ON s.id = ct_counts.service_id
		LEFT JOIN (
			SELECT service_id, SUM(daily_up) as today_up, SUM(daily_down) as today_down FROM inbound_traffic_history WHERE date = DATE('now', 'localtime') GROUP BY service_id
		) today_traffic ON s.id = today_traffic.service_id
		ORDER BY
			s.last_seen DESC;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int
		var ipAddress, lastSeen, status string
		var customName sql.NullString
		var inboundCount, clientCount int
		var todayInboundUp, todayInboundDown int64

		err := rows.Scan(&id, &ipAddress, &customName, &lastSeen, &status, &inboundCount, &clientCount, &todayInboundUp, &todayInboundDown)
		if err != nil {
			return nil, err
		}
		result := map[string]interface{}{
			"id":                 id,
			"ip":                 ipAddress,
			"custom_name":        customName.String,
			"last_seen":          lastSeen,
			"status":             status,
			"inbound_count":      inboundCount,
			"client_count":       clientCount,
			"today_inbound_up":   todayInboundUp,
			"today_inbound_down": todayInboundDown,
		}
		results = append(results, result)
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
		SELECT id, ip_address, custom_name, first_seen, last_seen, status
		FROM services WHERE id = ?
	`, serviceID).Scan(&service.ID, &rawIPAddress, &customName,
		&service.FirstSeen, &service.LastSeen, &service.Status)
	if err != nil {
		return nil, err
	}
	service.IPAddress = rawIPAddress

	// 批量查询所有入站端口的今日流量
	inboundTrafficMap := make(map[int]struct{ Up, Down int64 })
	inboundTrafficRows, err := d.db.Query(`
		SELECT inbound_traffic_id, COALESCE(daily_up,0), COALESCE(daily_down,0)
		FROM inbound_traffic_history
		WHERE service_id = ? AND date = DATE('now', 'localtime')
	`, serviceID)
	if err == nil {
		defer inboundTrafficRows.Close()
		for inboundTrafficRows.Next() {
			var inboundID int
			var up, down int64
			if err := inboundTrafficRows.Scan(&inboundID, &up, &down); err == nil {
				inboundTrafficMap[inboundID] = struct{ Up, Down int64 }{up, down}
			}
		}
	}

	// 获取入站流量（基础信息）
	inboundRows, err := d.db.Query(`
		SELECT id, service_id, tag, port, custom_name, last_updated, status
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
			&record.LastUpdated, &record.Status)
		if err != nil {
			return nil, err
		}
		record.CustomName = customName.String
		// 从 map 获取今日流量
		if v, ok := inboundTrafficMap[record.ID]; ok {
			record.Up = v.Up
			record.Down = v.Down
		} else {
			record.Up = 0
			record.Down = 0
		}
		inboundTraffics = append(inboundTraffics, record)
	}

	// 批量查询所有客户端的今日流量
	clientTrafficMap := make(map[int]struct{ Up, Down int64 })
	clientTrafficRows, err := d.db.Query(`
		SELECT client_traffic_id, COALESCE(daily_up,0), COALESCE(daily_down,0)
		FROM client_traffic_history
		WHERE service_id = ? AND date = DATE('now', 'localtime')
	`, serviceID)
	if err == nil {
		defer clientTrafficRows.Close()
		for clientTrafficRows.Next() {
			var clientID int
			var up, down int64
			if err := clientTrafficRows.Scan(&clientID, &up, &down); err == nil {
				clientTrafficMap[clientID] = struct{ Up, Down int64 }{up, down}
			}
		}
	}

	// 获取客户端流量（基础信息）
	clientRows, err := d.db.Query(`
		SELECT id, service_id, email, custom_name, last_updated, status
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
		err := clientRows.Scan(&record.ID, &record.ServiceID, &record.Email, &customName, &record.LastUpdated, &record.Status)
		if err != nil {
			return nil, err
		}
		record.CustomName = customName.String
		// 从 map 获取今日流量
		if v, ok := clientTrafficMap[record.ID]; ok {
			record.Up = v.Up
			record.Down = v.Down
		} else {
			record.Up = 0
			record.Down = 0
		}
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

	// 删除服务记录
	_, err = tx.Exec("DELETE FROM services WHERE id = ?", serviceID)
	if err != nil {
		return fmt.Errorf("删除服务记录失败: %v", err)
	}

	log.Printf("服务ID %d 删除成功", serviceID)
	return tx.Commit()
}

// 通用：处理每日流量统计
func (d *Database) processDailyTraffic(tx *sql.Tx, table string, historyTable string, idField string, extraField string) error {
	var query string
	if table == "inbound_traffics" {
		query = `SELECT id, service_id, tag, up, down FROM inbound_traffics WHERE status = 'active'`
	} else if table == "client_traffics" {
		query = `SELECT id, service_id, email, up, down FROM client_traffics WHERE status = 'active'`
	} else {
		return fmt.Errorf("不支持的表: %s", table)
	}
	rows, err := tx.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id, serviceID int
		var extra string
		var dailyUp, dailyDown int64
		err := rows.Scan(&id, &serviceID, &extra, &dailyUp, &dailyDown)
		if err != nil {
			return err
		}
		if dailyUp > 0 || dailyDown > 0 {
			var insertQuery string
			if table == "inbound_traffics" {
				insertQuery = `INSERT OR REPLACE INTO inbound_traffic_history (inbound_traffic_id, service_id, tag, date, daily_up, daily_down, created_at) VALUES (?, ?, ?, DATE('now', 'localtime'), ?, ?, CURRENT_TIMESTAMP)`
				_, err = tx.Exec(insertQuery, id, serviceID, extra, dailyUp, dailyDown)
			} else {
				insertQuery = `INSERT OR REPLACE INTO client_traffic_history (client_traffic_id, service_id, email, date, daily_up, daily_down, created_at) VALUES (?, ?, ?, DATE('now', 'localtime'), ?, ?, CURRENT_TIMESTAMP)`
				_, err = tx.Exec(insertQuery, id, serviceID, extra, dailyUp, dailyDown)
			}
			if err != nil {
				return err
			}
		}
		// 清零今日流量
		_, err = tx.Exec("UPDATE "+table+" SET up = 0, down = 0 WHERE id = ?", id)
		if err != nil {
			return err
		}
	}
	return nil
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
	err = d.processDailyTraffic(tx, "inbound_traffics", "inbound_traffic_history", "tag", "tag")
	if err != nil {
		return err
	}

	// 处理客户端流量记录
	err = d.processDailyTraffic(tx, "client_traffics", "client_traffic_history", "email", "email")
	if err != nil {
		return err
	}

	log.Println("每日流量统计完成")
	return tx.Commit()
}

// 创建/更新hy2配置表（支持多条配置）
func (d *Database) InitHy2ConfigTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS hy2_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_api_password TEXT NOT NULL DEFAULT '',
		source_api_host TEXT NOT NULL DEFAULT '',
		source_api_port TEXT NOT NULL DEFAULT '',
		target_api_url TEXT NOT NULL DEFAULT ''
	);
	`
	_, err := d.db.Exec(sql)
	return err
}

// 获取全部hy2配置
func (d *Database) GetAllHy2Configs() ([]Hy2Config, error) {
	rows, err := d.db.Query(`SELECT id, source_api_password, source_api_host, source_api_port, target_api_url FROM hy2_config`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var configs []Hy2Config
	for rows.Next() {
		var cfg Hy2Config
		err := rows.Scan(&cfg.ID, &cfg.SourceAPIPassword, &cfg.SourceAPIHost, &cfg.SourceAPIPort, &cfg.TargetAPIURL)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

// 新增hy2配置
func (d *Database) AddHy2Config(cfg *Hy2Config) error {
	_, err := d.db.Exec(`INSERT INTO hy2_config (source_api_password, source_api_host, source_api_port, target_api_url) VALUES (?, ?, ?, ?)`,
		cfg.SourceAPIPassword, cfg.SourceAPIHost, cfg.SourceAPIPort, cfg.TargetAPIURL)
	return err
}

// 更新hy2配置
func (d *Database) UpdateHy2Config(cfg *Hy2Config) error {
	_, err := d.db.Exec(`UPDATE hy2_config SET source_api_password=?, source_api_host=?, source_api_port=?, target_api_url=? WHERE id=?`,
		cfg.SourceAPIPassword, cfg.SourceAPIHost, cfg.SourceAPIPort, cfg.TargetAPIURL, cfg.ID)
	return err
}

// 删除hy2配置
func (d *Database) DeleteHy2Config(id int) error {
	_, err := d.db.Exec(`DELETE FROM hy2_config WHERE id=?`, id)
	return err
}

// 删除全部hy2配置
func (d *Database) DeleteAllHy2Configs() error {
	_, err := d.db.Exec("DELETE FROM hy2_config")
	return err
}
