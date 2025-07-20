package database

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// 数据库API处理器
type DatabaseAPI struct {
	db *Database
}

// 创建数据库API处理器
func NewDatabaseAPI(db *Database) *DatabaseAPI {
	return &DatabaseAPI{db: db}
}

// 注册API路由
func (api *DatabaseAPI) RegisterRoutes(r *gin.Engine) {
	// 数据库相关API（需要认证）
	dbGroup := r.Group("/api/db")
	dbGroup.Use(AuthMiddleware()) // 添加认证中间件
	{
		// 服务管理
		dbGroup.GET("/services", api.GetServices)
		dbGroup.GET("/services/:id", api.GetServiceDetail)
		dbGroup.GET("/services/:id/traffic", api.GetServiceTraffic)
		dbGroup.DELETE("/services/:id", api.DeleteService)

		// 流量统计
		dbGroup.GET("/traffic/summary", api.GetTrafficSummary)
		dbGroup.GET("/traffic/history", api.GetTrafficHistory)
		dbGroup.GET("/traffic/weekly/:service_id", api.GetWeeklyTraffic)
		dbGroup.GET("/traffic/monthly/:service_id", api.GetMonthlyTraffic)

		// 原始数据
		dbGroup.GET("/raw-requests", api.GetRawRequests)
		dbGroup.GET("/raw-requests/:id", api.GetRawRequestDetail)

		// 手动触发每日统计
		dbGroup.POST("/daily-summary", api.TriggerDailySummary)

		// 端口和用户详情
		dbGroup.GET("/port-detail/:service_id/:tag", api.GetPortDetail)
		dbGroup.GET("/user-detail/:service_id/:email", api.GetUserDetail)

		// 自定义名称管理
		dbGroup.PUT("/services/:id/custom-name", api.UpdateServiceCustomName)
		dbGroup.PUT("/inbound/:service_id/:tag/custom-name", api.UpdateInboundCustomName)
		dbGroup.PUT("/client/:service_id/:email/custom-name", api.UpdateClientCustomName)
	}
}

// 获取所有服务列表
func (api *DatabaseAPI) GetServices(c *gin.Context) {
	services, err := api.db.GetServiceSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取服务列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取服务列表成功",
		"data":    services,
	})
}

// 获取服务详情 - 重定向到GetServiceTraffic
func (api *DatabaseAPI) GetServiceDetail(c *gin.Context) {
	api.GetServiceTraffic(c)
}

// 获取服务流量详情
func (api *DatabaseAPI) GetServiceTraffic(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的服务ID",
		})
		return
	}

	traffic, err := api.db.GetServiceTraffic(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取服务流量失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取服务流量成功",
		"data":    traffic,
	})
}

// 删除服务及其所有相关数据
func (api *DatabaseAPI) DeleteService(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的服务ID",
		})
		return
	}

	err = api.db.DeleteService(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "删除服务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "服务删除成功",
		"data": gin.H{
			"deleted_service_id": id,
		},
	})
}

// 获取流量汇总
func (api *DatabaseAPI) GetTrafficSummary(c *gin.Context) {
	services, err := api.db.GetServiceSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "获取流量汇总失败: " + err.Error(),
		})
		return
	}

	// 计算总流量
	var totalUp, totalDown int64
	for _, service := range services {
		if up, ok := service["total_inbound_up"].(int64); ok {
			totalUp += up
		}
		if down, ok := service["total_inbound_down"].(int64); ok {
			totalDown += down
		}
	}

	summary := gin.H{
		"total_services": len(services),
		"total_up":       totalUp,
		"total_down":     totalDown,
		"total_traffic":  totalUp + totalDown,
		"services":       services,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取流量汇总成功",
		"data":    summary,
	})
}

// 获取流量历史记录
func (api *DatabaseAPI) GetTrafficHistory(c *gin.Context) {
	// 获取查询参数
	serviceID := c.Query("service_id")
	tag := c.Query("tag")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// 构建查询条件
	query := `
		SELECT 
			ith.date,
			ith.tag,
			s.ip_address,
			ith.daily_up,
			ith.daily_down,
			ith.daily_up + ith.daily_down as total_daily
		FROM inbound_traffic_history ith
		JOIN services s ON ith.service_id = s.id
		WHERE 1=1
	`
	args := []interface{}{}

	if serviceID != "" {
		query += " AND ith.service_id = ?"
		args = append(args, serviceID)
	}

	if tag != "" {
		query += " AND ith.tag = ?"
		args = append(args, tag)
	}

	if startDate != "" {
		query += " AND ith.date >= ?"
		args = append(args, startDate)
	}

	if endDate != "" {
		query += " AND ith.date <= ?"
		args = append(args, endDate)
	}

	query += " ORDER BY ith.date DESC, ith.tag"

	rows, err := api.db.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询流量历史失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var date, tag, ipAddress string
		var dailyUp, dailyDown, totalDaily int64

		err := rows.Scan(&date, &tag, &ipAddress, &dailyUp, &dailyDown, &totalDaily)
		if err != nil {
			continue
		}

		record := map[string]interface{}{
			"date":        date,
			"tag":         tag,
			"ip_address":  api.db.maskIPAddress(ipAddress),
			"daily_up":    dailyUp,
			"daily_down":  dailyDown,
			"total_daily": totalDaily,
		}
		history = append(history, record)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取流量历史成功",
		"data":    history,
	})
}

// 获取原始请求记录
func (api *DatabaseAPI) GetRawRequests(c *gin.Context) {
	// 获取查询参数
	serviceID := c.Query("service_id")
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	query := `
		SELECT 
			rr.id,
			rr.service_id,
			rr.client_ip,
			rr.user_agent,
			rr.request_body,
			rr.received_at,
			rr.processed,
			s.ip_address as service_ip
		FROM raw_requests rr
		JOIN services s ON rr.service_id = s.id
		WHERE 1=1
	`
	args := []interface{}{}

	if serviceID != "" {
		query += " AND rr.service_id = ?"
		args = append(args, serviceID)
	}

	query += " ORDER BY rr.received_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := api.db.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询原始请求失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var requests []map[string]interface{}
	for rows.Next() {
		var id, serviceID int
		var clientIP, userAgent, requestBody, receivedAt, serviceIP string
		var processed bool

		err := rows.Scan(&id, &serviceID, &clientIP, &userAgent, &requestBody, &receivedAt, &processed, &serviceIP)
		if err != nil {
			continue
		}

		// 尝试解析请求体为JSON
		var parsedBody interface{}
		json.Unmarshal([]byte(requestBody), &parsedBody)

		record := map[string]interface{}{
			"id":           id,
			"service_id":   serviceID,
			"service_ip":   api.db.maskIPAddress(serviceIP),
			"client_ip":    clientIP,
			"user_agent":   userAgent,
			"request_body": parsedBody,
			"received_at":  receivedAt,
			"processed":    processed,
		}
		requests = append(requests, record)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取原始请求成功",
		"data":    requests,
	})
}

// 获取原始请求详情
func (api *DatabaseAPI) GetRawRequestDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的请求ID",
		})
		return
	}

	query := `
		SELECT 
			rr.id,
			rr.service_id,
			rr.client_ip,
			rr.user_agent,
			rr.request_body,
			rr.parsed_data,
			rr.received_at,
			rr.processed,
			s.ip_address as service_ip
		FROM raw_requests rr
		JOIN services s ON rr.service_id = s.id
		WHERE rr.id = ?
	`

	var serviceID int
	var clientIP, userAgent, requestBody, parsedData, receivedAt, serviceIP string
	var processed bool

	err = api.db.db.QueryRow(query, id).Scan(&id, &serviceID, &clientIP, &userAgent, &requestBody, &parsedData, &receivedAt, &processed, &serviceIP)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "请求记录不存在",
		})
		return
	}

	// 解析JSON数据
	var parsedBody, parsedDataObj interface{}
	json.Unmarshal([]byte(requestBody), &parsedBody)
	json.Unmarshal([]byte(parsedData), &parsedDataObj)

	record := map[string]interface{}{
		"id":           id,
		"service_id":   serviceID,
		"service_ip":   serviceIP,
		"client_ip":    clientIP,
		"user_agent":   userAgent,
		"request_body": parsedBody,
		"parsed_data":  parsedDataObj,
		"received_at":  receivedAt,
		"processed":    processed,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取请求详情成功",
		"data":    record,
	})
}

// 获取服务7天流量数据
func (api *DatabaseAPI) GetWeeklyTraffic(c *gin.Context) {
	serviceIDStr := c.Param("service_id")
	serviceID, err := strconv.Atoi(serviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的服务ID",
		})
		return
	}

	// 获取过去7天的日期（从6天前到今天，今天在最右边）
	dates := make([]string, 7)
	trafficData := make(map[string]map[string]int64)

	// 直接生成正确的顺序：6天前, 5天前, ..., 昨天, 今天
	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, -(6 - i))
		dateStr := date.Format("2006-01-02")
		dates[i] = dateStr
		trafficData[dateStr] = map[string]int64{
			"upload":   0,
			"download": 0,
		}
	}

	// 查询历史流量数据（前6天）
	historyQuery := `
		SELECT 
			ith.date,
			SUM(ith.daily_up) as total_up,
			SUM(ith.daily_down) as total_down
		FROM inbound_traffic_history ith
		WHERE ith.service_id = ? AND ith.date >= DATE('now', '-6 days') AND ith.date < DATE('now')
		GROUP BY ith.date
		ORDER BY ith.date
	`

	historyRows, err := api.db.db.Query(historyQuery, serviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询历史流量失败: " + err.Error(),
		})
		return
	}
	defer historyRows.Close()

	for historyRows.Next() {
		var date string
		var totalUp, totalDown int64

		err := historyRows.Scan(&date, &totalUp, &totalDown)
		if err != nil {
			continue
		}

		// 处理日期格式，去掉时间部分
		dateOnly := date
		if len(date) > 10 {
			dateOnly = date[:10]
		}

		if data, exists := trafficData[dateOnly]; exists {
			data["upload"] = totalUp
			data["download"] = totalDown
		}
	}

	// 查询今日实时流量数据
	todayQuery := `
		SELECT 
			SUM(it.up) as total_up,
			SUM(it.down) as total_down
		FROM inbound_traffics it
		WHERE it.service_id = ? AND it.status = 'active'
	`

	var todayUp, todayDown int64
	err = api.db.db.QueryRow(todayQuery, serviceID).Scan(&todayUp, &todayDown)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询今日流量失败: " + err.Error(),
		})
		return
	}

	// 设置今日数据
	today := time.Now().Format("2006-01-02")
	if data, exists := trafficData[today]; exists {
		data["upload"] = todayUp
		data["download"] = todayDown
	}

	// 构建响应数据
	uploadData := make([]int64, 7)
	downloadData := make([]int64, 7)

	for i, date := range dates {
		if data, exists := trafficData[date]; exists {
			uploadData[i] = data["upload"]
			downloadData[i] = data["download"]
		}
	}

	result := gin.H{
		"dates":         dates,
		"upload_data":   uploadData,
		"download_data": downloadData,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取7天流量数据成功",
		"data":    result,
	})
}

// 获取服务30天流量数据
func (api *DatabaseAPI) GetMonthlyTraffic(c *gin.Context) {
	serviceIDStr := c.Param("service_id")
	serviceID, err := strconv.Atoi(serviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的服务ID",
		})
		return
	}

	// 获取过去30天的日期（从29天前到今天，今天在最右边）
	dates := make([]string, 30)
	trafficData := make(map[string]map[string]int64)

	// 直接生成正确的顺序：29天前, 28天前, ..., 昨天, 今天
	for i := 0; i < 30; i++ {
		date := time.Now().AddDate(0, 0, -(29 - i))
		dateStr := date.Format("2006-01-02")
		dates[i] = dateStr
		trafficData[dateStr] = map[string]int64{
			"upload":   0,
			"download": 0,
		}
	}

	// 查询历史流量数据（前29天）
	historyQuery := `
		SELECT 
			ith.date,
			SUM(ith.daily_up) as total_up,
			SUM(ith.daily_down) as total_down
		FROM inbound_traffic_history ith
		WHERE ith.service_id = ? AND ith.date >= DATE('now', '-29 days') AND ith.date < DATE('now')
		GROUP BY ith.date
		ORDER BY ith.date
	`

	historyRows, err := api.db.db.Query(historyQuery, serviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询历史流量失败: " + err.Error(),
		})
		return
	}
	defer historyRows.Close()

	for historyRows.Next() {
		var date string
		var totalUp, totalDown int64

		err := historyRows.Scan(&date, &totalUp, &totalDown)
		if err != nil {
			continue
		}

		// 处理日期格式，去掉时间部分
		dateOnly := date
		if len(date) > 10 {
			dateOnly = date[:10]
		}

		if data, exists := trafficData[dateOnly]; exists {
			data["upload"] = totalUp
			data["download"] = totalDown
		}
	}

	// 查询今日实时流量数据
	todayQuery := `
		SELECT 
			SUM(it.up) as total_up,
			SUM(it.down) as total_down
		FROM inbound_traffics it
		WHERE it.service_id = ? AND it.status = 'active'
	`

	var todayUp, todayDown int64
	err = api.db.db.QueryRow(todayQuery, serviceID).Scan(&todayUp, &todayDown)
	if err != nil && err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询今日流量失败: " + err.Error(),
		})
		return
	}

	// 设置今日数据
	today := time.Now().Format("2006-01-02")
	if data, exists := trafficData[today]; exists {
		data["upload"] = todayUp
		data["download"] = todayDown
	}

	// 构建响应数据
	uploadData := make([]int64, 30)
	downloadData := make([]int64, 30)

	for i, date := range dates {
		if data, exists := trafficData[date]; exists {
			uploadData[i] = data["upload"]
			downloadData[i] = data["download"]
		}
	}

	result := gin.H{
		"dates":         dates,
		"upload_data":   uploadData,
		"download_data": downloadData,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取30天流量数据成功",
		"data":    result,
	})
}

// 手动触发每日统计
func (api *DatabaseAPI) TriggerDailySummary(c *gin.Context) {
	err := api.db.DailyTrafficSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "执行每日统计失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "每日统计执行成功",
		"data": gin.H{
			"executed_at": time.Now().Format("2006-01-02 15:04:05"),
		},
	})
}

// 获取端口详细流量信息
func (api *DatabaseAPI) GetPortDetail(c *gin.Context) {
	serviceIDStr := c.Param("service_id")
	tag := c.Param("tag")

	if serviceIDStr == "" || tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少服务ID或端口标签参数",
		})
		return
	}

	serviceID, err := strconv.Atoi(serviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的服务ID",
		})
		return
	}

	// 获取端口基本信息
	var ipAddress, serviceName, tagName string
	var port int
	var totalUp, totalDown int64
	var lastSeen string

	portQuery := `
		SELECT 
			s.ip_address,
			s.service_name,
			it.tag,
			it.port,
			COALESCE(SUM(ith.daily_up), 0) as total_up,
			COALESCE(SUM(ith.daily_down), 0) as total_down,
			it.last_updated as last_seen
		FROM inbound_traffics it
		JOIN services s ON it.service_id = s.id
		LEFT JOIN inbound_traffic_history ith ON it.id = ith.inbound_traffic_id
		WHERE it.service_id = ? AND it.tag = ?
		GROUP BY s.ip_address, s.service_name, it.tag, it.port, it.last_updated
	`

	err = api.db.db.QueryRow(portQuery, serviceID, tag).Scan(
		&ipAddress, &serviceName, &tagName, &port, &totalUp, &totalDown, &lastSeen,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "端口信息不存在",
		})
		return
	}

	// 检查端口是否真正活跃（最近有流量）
	isActive := false
	if lastSeen != "" {
		// 检查最后活跃时间是否在60秒内，并且有实际流量
		var lastSeenTime time.Time
		timeFormats := []string{
			"2006-01-02 15:04:05",
			time.RFC3339,
			"2006-01-02T15:04:05Z",
		}

		for _, format := range timeFormats {
			if t, err := time.Parse(format, lastSeen); err == nil {
				lastSeenTime = t
				break
			}
		}

		if !lastSeenTime.IsZero() {
			timeDiff := time.Since(lastSeenTime).Seconds()
			isActive = timeDiff <= 60 && (totalUp > 0 || totalDown > 0)
		}
	}

	portInfo := map[string]interface{}{
		"ip_address":   api.db.maskIPAddress(ipAddress),
		"service_name": serviceName,
		"tag":          tagName,
		"port":         port,
		"total_up":     totalUp,
		"total_down":   totalDown,
		"last_seen":    lastSeen,
		"is_active":    isActive,
	}

	// 获取端口历史流量数据
	// 获取days参数
	days := 7
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 && v <= 30 {
			days = v
		}
	}

	historyQuery := `
		SELECT 
			date,
			daily_up,
			daily_down,
			daily_up + daily_down as total_daily
		FROM inbound_traffic_history
		WHERE service_id = ? AND tag = ?
		ORDER BY date DESC
		LIMIT ?
	`

	rows, err := api.db.db.Query(historyQuery, serviceID, tag, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询历史数据失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var date string
		var dailyUp, dailyDown, totalDaily int64

		err := rows.Scan(&date, &dailyUp, &dailyDown, &totalDaily)
		if err != nil {
			continue
		}

		record := map[string]interface{}{
			"date":        date,
			"daily_up":    dailyUp,
			"daily_down":  dailyDown,
			"total_daily": totalDaily,
		}
		history = append(history, record)
	}

	result := gin.H{
		"port_info": portInfo,
		"history":   history,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取端口详情成功",
		"data":    result,
	})
}

// 获取用户详细流量信息
func (api *DatabaseAPI) GetUserDetail(c *gin.Context) {
	serviceID := c.Param("service_id")
	email := c.Param("email")

	if serviceID == "" || email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "缺少服务ID或用户邮箱参数",
		})
		return
	}

	// 获取用户基本信息
	var ipAddress, serviceName, userEmail, inboundTag string
	var totalUp, totalDown int64
	var lastSeen string

	userQuery := `
		SELECT 
			s.ip_address,
			s.service_name,
			ct.email,
			'' as inbound_tag,
			COALESCE(SUM(cth.daily_up), 0) as total_up,
			COALESCE(SUM(cth.daily_down), 0) as total_down,
			ct.last_updated as last_seen
		FROM client_traffics ct
		JOIN services s ON ct.service_id = s.id
		LEFT JOIN client_traffic_history cth ON ct.id = cth.client_traffic_id
		WHERE ct.service_id = ? AND ct.email = ?
		GROUP BY s.ip_address, s.service_name, ct.email, ct.last_updated
	`

	err := api.db.db.QueryRow(userQuery, serviceID, email).Scan(
		&ipAddress, &serviceName, &userEmail, &inboundTag, &totalUp, &totalDown, &lastSeen,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "用户信息不存在",
		})
		return
	}

	userInfo := map[string]interface{}{
		"ip_address":   api.db.maskIPAddress(ipAddress),
		"service_name": serviceName,
		"email":        userEmail,
		"inbound_tag":  inboundTag,
		"total_up":     totalUp,
		"total_down":   totalDown,
		"last_seen":    lastSeen,
	}

	// 获取用户历史流量数据
	// 获取days参数
	days := 7
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 && v <= 30 {
			days = v
		}
	}

	historyQuery := `
		SELECT 
			cth.date,
			cth.daily_up,
			cth.daily_down,
			cth.daily_up + cth.daily_down as total_daily
		FROM client_traffic_history cth
		JOIN client_traffics ct ON cth.client_traffic_id = ct.id
		WHERE ct.service_id = ? AND ct.email = ?
		ORDER BY cth.date DESC
		LIMIT ?
	`

	// 转换serviceID为整数
	serviceIDInt, err := strconv.Atoi(serviceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的服务ID",
		})
		return
	}

	rows, err := api.db.db.Query(historyQuery, serviceIDInt, email, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询历史数据失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var history []map[string]interface{}
	rowCount := 0
	for rows.Next() {
		var date string
		var dailyUp, dailyDown, totalDaily int64

		err := rows.Scan(&date, &dailyUp, &dailyDown, &totalDaily)
		if err != nil {
			continue
		}

		record := map[string]interface{}{
			"date":        date,
			"daily_up":    dailyUp,
			"daily_down":  dailyDown,
			"total_daily": totalDaily,
		}
		history = append(history, record)
		rowCount++
	}

	// 检查是否有错误
	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "遍历历史数据失败: " + err.Error(),
		})
		return
	}

	// 添加调试信息
	result := gin.H{
		"user_info": userInfo,
		"history":   history,
		"debug": gin.H{
			"service_id":     serviceID,
			"service_id_int": serviceIDInt,
			"email":          email,
			"row_count":      rowCount,
			"history_length": len(history),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取用户详情成功",
		"data":    result,
	})
}

// 更新服务自定义名称
func (api *DatabaseAPI) UpdateServiceCustomName(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的服务ID",
		})
		return
	}

	var request struct {
		CustomName string `json:"custom_name"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 更新自定义名称
	_, err = api.db.db.Exec("UPDATE services SET custom_name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", request.CustomName, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新服务名称失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "服务名称更新成功",
		"data": gin.H{
			"service_id":  id,
			"custom_name": request.CustomName,
		},
	})
}

// 更新入站端口自定义名称
func (api *DatabaseAPI) UpdateInboundCustomName(c *gin.Context) {
	serviceIDStr := c.Param("service_id")
	tag := c.Param("tag")

	serviceID, err := strconv.Atoi(serviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的服务ID",
		})
		return
	}

	var request struct {
		CustomName string `json:"custom_name"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 更新自定义名称
	_, err = api.db.db.Exec("UPDATE inbound_traffics SET custom_name = ?, updated_at = CURRENT_TIMESTAMP WHERE service_id = ? AND tag = ?", request.CustomName, serviceID, tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新端口名称失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "端口名称更新成功",
		"data": gin.H{
			"service_id":  serviceID,
			"tag":         tag,
			"custom_name": request.CustomName,
		},
	})
}

// 更新客户端自定义名称
func (api *DatabaseAPI) UpdateClientCustomName(c *gin.Context) {
	serviceIDStr := c.Param("service_id")
	email := c.Param("email")

	serviceID, err := strconv.Atoi(serviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的服务ID",
		})
		return
	}

	var request struct {
		CustomName string `json:"custom_name"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 更新自定义名称
	_, err = api.db.db.Exec("UPDATE client_traffics SET custom_name = ?, updated_at = CURRENT_TIMESTAMP WHERE service_id = ? AND email = ?", request.CustomName, serviceID, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "更新用户名称失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户名称更新成功",
		"data": gin.H{
			"service_id":  serviceID,
			"email":       email,
			"custom_name": request.CustomName,
		},
	})
}
