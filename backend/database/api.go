package database

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
		dbGroup.GET("/services/:id", api.GetServiceTraffic) // 直接使用GetServiceTraffic
		dbGroup.GET("/services/:id/traffic", api.GetServiceTraffic)
		dbGroup.DELETE("/services/:id", api.DeleteService)

		// 流量统计
		dbGroup.GET("/traffic/history", api.GetTrafficHistory)
		dbGroup.GET("/traffic/weekly/:service_id", api.GetWeeklyTraffic)
		dbGroup.GET("/traffic/monthly/:service_id", api.GetMonthlyTraffic)

		// 手动触发每日统计
		dbGroup.POST("/daily-summary", api.TriggerDailySummary)

		// 端口和用户详情
		dbGroup.GET("/port-detail/:service_id/:tag", api.GetPortDetail)
		dbGroup.GET("/user-detail/:service_id/:email", api.GetUserDetail)

		// 自定义名称管理
		dbGroup.PUT("/services/:id/custom-name", api.UpdateServiceCustomName)
		dbGroup.PUT("/inbound/:service_id/:tag/custom-name", api.UpdateInboundCustomName)
		dbGroup.PUT("/client/:service_id/:email/custom-name", api.UpdateClientCustomName)

		// 下载历史数据
		dbGroup.GET("/download/port-history/:service_id/:tag", api.DownloadPortHistory)
		dbGroup.GET("/download/user-history/:service_id/:email", api.DownloadUserHistory)
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
			"ip_address":  ipAddress,
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
	for i := 0; i < 7; i++ {
		date := time.Now().In(time.Local).AddDate(0, 0, -(6 - i))
		dateStr := date.Format("2006-01-02")
		dates[i] = dateStr
		trafficData[dateStr] = map[string]int64{
			"upload":   0,
			"download": 0,
		}
	}

	// 查询7天内所有历史流量（包含今天），全部用 localtime
	historyQuery := `
		SELECT 
			ith.date,
			SUM(ith.daily_up) as total_up,
			SUM(ith.daily_down) as total_down
		FROM inbound_traffic_history ith
		WHERE ith.service_id = ? AND ith.date >= DATE('now', '-6 days', 'localtime') AND ith.date <= DATE('now', 'localtime')
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

	fmt.Println("开始遍历 historyRows")
	count := 0
	for historyRows.Next() {
		count++
		var date string
		var totalUp, totalDown int64
		err := historyRows.Scan(&date, &totalUp, &totalDown)
		if err != nil {
			continue
		}
		fmt.Printf("数据库查到的date: '%s', totalUp: %d, totalDown: %d\n", date, totalUp, totalDown)
		date = strings.TrimSpace(date)
		if len(date) > 10 {
			date = date[:10]
		}
		// 兼容数据库查出来的 date 和 dates 数组格式
		found := false
		for _, d := range dates {
			if strings.TrimSpace(d) == date {
				trafficData[d]["upload"] = totalUp
				trafficData[d]["download"] = totalDown
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("未命中trafficData，date='%s', dates=%v\n", date, dates)
		}
	}
	fmt.Printf("historyRows 总数: %d\n", count)
	fmt.Printf("最终 trafficData: %+v\n", trafficData)

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
	for i := 0; i < 30; i++ {
		date := time.Now().In(time.Local).AddDate(0, 0, -(29 - i))
		dateStr := date.Format("2006-01-02")
		dates[i] = dateStr
		trafficData[dateStr] = map[string]int64{
			"upload":   0,
			"download": 0,
		}
	}

	// 查询30天内所有历史流量（包含今天），全部用 localtime
	historyQuery := `
		SELECT 
			ith.date,
			SUM(ith.daily_up) as total_up,
			SUM(ith.daily_down) as total_down
		FROM inbound_traffic_history ith
		WHERE ith.service_id = ? AND ith.date >= DATE('now', '-29 days', 'localtime') AND ith.date <= DATE('now', 'localtime')
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

	fmt.Println("开始遍历 historyRows")
	count := 0
	for historyRows.Next() {
		count++
		var date string
		var totalUp, totalDown int64
		err := historyRows.Scan(&date, &totalUp, &totalDown)
		if err != nil {
			continue
		}
		fmt.Printf("数据库查到的date: '%s', totalUp: %d, totalDown: %d\n", date, totalUp, totalDown)
		date = strings.TrimSpace(date)
		if len(date) > 10 {
			date = date[:10]
		}
		// 兼容数据库查出来的 date 和 dates 数组格式
		found := false
		for _, d := range dates {
			if strings.TrimSpace(d) == date {
				trafficData[d]["upload"] = totalUp
				trafficData[d]["download"] = totalDown
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("未命中trafficData，date='%s', dates=%v\n", date, dates)
		}
	}
	fmt.Printf("historyRows 总数: %d\n", count)
	fmt.Printf("最终 trafficData: %+v\n", trafficData)

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
			"executed_at": time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
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
	var customName sql.NullString
	var port int
	var lastSeen string
	portQuery := `
		SELECT 
			s.ip_address,
			s.service_name,
			it.tag,
			it.port,
			it.last_updated as last_seen,
			it.custom_name
		FROM inbound_traffics it
		JOIN services s ON it.service_id = s.id
		WHERE it.service_id = ? AND it.tag = ?
	`
	err = api.db.db.QueryRow(portQuery, serviceID, tag).Scan(
		&ipAddress, &serviceName, &tagName, &port, &lastSeen, &customName,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "端口信息不存在: " + err.Error(),
		})
		return
	}

	// 查询今日流量
	var currentUp, currentDown int64
	err = api.db.db.QueryRow(`SELECT COALESCE(daily_up,0), COALESCE(daily_down,0) FROM inbound_traffic_history WHERE service_id = ? AND tag = ? AND date = DATE('now', 'localtime')`, serviceID, tag).Scan(&currentUp, &currentDown)
	if err != nil && err != sql.ErrNoRows {
		currentUp, currentDown = 0, 0
	}

	// 计算历史累计流量
	var totalUp, totalDown int64
	historySumQuery := `
		SELECT 
			COALESCE(SUM(daily_up), 0) as total_up,
			COALESCE(SUM(daily_down), 0) as total_down
		FROM inbound_traffic_history
		WHERE service_id = ? AND tag = ?
	`
	err = api.db.db.QueryRow(historySumQuery, serviceID, tag).Scan(&totalUp, &totalDown)
	if err != nil {
		totalUp = currentUp
		totalDown = currentDown
	}

	// 检查端口是否真正活跃（最近有流量）
	isActive := false
	if lastSeen != "" {
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
			isActive = timeDiff <= 60 && (currentUp > 0 || currentDown > 0)
		}
	}

	portInfo := map[string]interface{}{
		"ip_address":   ipAddress,
		"service_name": serviceName,
		"tag":          tagName,
		"port":         port,
		"total_up":     totalUp,
		"total_down":   totalDown,
		"current_up":   currentUp,
		"current_down": currentDown,
		"last_seen":    lastSeen,
		"is_active":    isActive,
		"custom_name":  customName.String,
	}

	// 获取days参数，默认7天
	days := 7
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 && v <= 30 {
			days = v
		}
	}

	// 构造最近days天的日期数组
	dates := make([]string, days)
	for i := 0; i < days; i++ {
		date := time.Now().In(time.Local).AddDate(0, 0, -(days - 1 - i))
		dateStr := date.Format("2006-01-02")
		dates[i] = dateStr
	}

	// 查询历史流量，补全为0
	historyQuery := `
		SELECT date, daily_up, daily_down, daily_up + daily_down as total_daily
		FROM inbound_traffic_history
		WHERE service_id = ? AND tag = ? AND date >= DATE('now', ? || ' days', 'localtime') AND date <= DATE('now', 'localtime')
	`
	rows, err := api.db.db.Query(historyQuery, serviceID, tag, fmt.Sprintf("-%d", days-1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询历史数据失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	historyMap := make(map[string]map[string]int64)
	for rows.Next() {
		var date string
		var dailyUp, dailyDown, totalDaily int64
		err := rows.Scan(&date, &dailyUp, &dailyDown, &totalDaily)
		if err != nil {
			continue
		}
		if len(date) > 10 {
			date = date[:10]
		}
		historyMap[date] = map[string]int64{
			"daily_up":    dailyUp,
			"daily_down":  dailyDown,
			"total_daily": totalDaily,
		}
	}

	history := make([]map[string]interface{}, days)
	for i, d := range dates {
		item := map[string]interface{}{
			"date":        d,
			"daily_up":    int64(0),
			"daily_down":  int64(0),
			"total_daily": int64(0),
		}
		if v, ok := historyMap[d]; ok {
			item["daily_up"] = v["daily_up"]
			item["daily_down"] = v["daily_down"]
			item["total_daily"] = v["total_daily"]
		}
		history[i] = item
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

	var ipAddress, serviceName, userEmail, inboundTag string
	var customName sql.NullString
	var lastSeen string
	var err error
	userQuery := `
		SELECT 
			s.ip_address,
			s.service_name,
			ct.email,
			'' as inbound_tag,
			ct.last_updated as last_seen,
			ct.custom_name
		FROM client_traffics ct
		JOIN services s ON ct.service_id = s.id
		WHERE ct.service_id = ? AND ct.email = ?
	`
	err = api.db.db.QueryRow(userQuery, serviceID, email).Scan(
		&ipAddress, &serviceName, &userEmail, &inboundTag, &lastSeen, &customName,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "用户信息不存在: " + err.Error(),
		})
		return
	}

	// 查询今日流量
	var currentUp, currentDown int64
	serviceIDInt, _ := strconv.Atoi(serviceID)
	err = api.db.db.QueryRow(`SELECT COALESCE(daily_up,0), COALESCE(daily_down,0) FROM client_traffic_history WHERE client_traffic_id = (SELECT id FROM client_traffics WHERE service_id = ? AND email = ?) AND date = DATE('now', 'localtime')`, serviceIDInt, email).Scan(&currentUp, &currentDown)
	if err != nil && err != sql.ErrNoRows {
		currentUp, currentDown = 0, 0
	}

	// 计算用户历史累计流量
	var totalUp, totalDown int64
	userHistorySumQuery := `
		SELECT 
			COALESCE(SUM(daily_up), 0) as total_up,
			COALESCE(SUM(daily_down), 0) as total_down
		FROM client_traffic_history
		WHERE service_id = ? AND email = ?
	`
	err = api.db.db.QueryRow(userHistorySumQuery, serviceID, email).Scan(&totalUp, &totalDown)
	if err != nil {
		totalUp = currentUp
		totalDown = currentDown
	}

	userInfo := map[string]interface{}{
		"ip_address":   ipAddress,
		"service_name": serviceName,
		"email":        userEmail,
		"inbound_tag":  inboundTag,
		"total_up":     totalUp,
		"total_down":   totalDown,
		"current_up":   currentUp,
		"current_down": currentDown,
		"last_seen":    lastSeen,
		"custom_name":  customName.String,
	}

	// 获取days参数，默认7天
	days := 7
	if d := c.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 && v <= 30 {
			days = v
		}
	}

	// 构造最近days天的日期数组
	dates := make([]string, days)
	for i := 0; i < days; i++ {
		date := time.Now().In(time.Local).AddDate(0, 0, -(days - 1 - i))
		dateStr := date.Format("2006-01-02")
		dates[i] = dateStr
	}

	// 查询历史流量，补全为0
	historyQuery := `
		SELECT date, daily_up, daily_down, daily_up + daily_down as total_daily
		FROM client_traffic_history
		WHERE service_id = ? AND email = ? AND date >= DATE('now', ? || ' days', 'localtime') AND date <= DATE('now', 'localtime')
	`
	rows, err := api.db.db.Query(historyQuery, serviceIDInt, email, fmt.Sprintf("-%d", days-1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "查询历史数据失败: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	historyMap := make(map[string]map[string]int64)
	for rows.Next() {
		var date string
		var dailyUp, dailyDown, totalDaily int64
		err := rows.Scan(&date, &dailyUp, &dailyDown, &totalDaily)
		if err != nil {
			continue
		}
		if len(date) > 10 {
			date = date[:10]
		}
		historyMap[date] = map[string]int64{
			"daily_up":    dailyUp,
			"daily_down":  dailyDown,
			"total_daily": totalDaily,
		}
	}

	history := make([]map[string]interface{}, days)
	for i, d := range dates {
		item := map[string]interface{}{
			"date":        d,
			"daily_up":    int64(0),
			"daily_down":  int64(0),
			"total_daily": int64(0),
		}
		if v, ok := historyMap[d]; ok {
			item["daily_up"] = v["daily_up"]
			item["daily_down"] = v["daily_down"]
			item["total_daily"] = v["total_daily"]
		}
		history[i] = item
	}

	result := gin.H{
		"user_info": userInfo,
		"history":   history,
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
			"error":   "更新入站失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "入站更新成功",
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

// 下载端口历史数据
func (api *DatabaseAPI) DownloadPortHistory(c *gin.Context) {
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

	portQuery := `
		SELECT 
			s.ip_address,
			s.service_name,
			it.tag,
			it.port
		FROM inbound_traffics it
		JOIN services s ON it.service_id = s.id
		WHERE it.service_id = ? AND it.tag = ?
	`

	err = api.db.db.QueryRow(portQuery, serviceID, tag).Scan(
		&ipAddress, &serviceName, &tagName, &port,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "端口信息不存在",
		})
		return
	}

	// 获取所有历史流量数据
	historyQuery := `
		SELECT 
			date,
			daily_up,
			daily_down,
			daily_up + daily_down as total_daily
		FROM inbound_traffic_history
		WHERE service_id = ? AND tag = ?
		ORDER BY date DESC
	`

	rows, err := api.db.db.Query(historyQuery, serviceID, tag)
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

	// 生成CSV内容
	csvContent := "日期,上传流量(Bytes),下载流量(Bytes),总流量(Bytes),上传流量(格式化),下载流量(格式化),总流量(格式化)\n"

	for _, record := range history {
		date := record["date"].(string)
		dailyUp := record["daily_up"].(int64)
		dailyDown := record["daily_down"].(int64)
		totalDaily := record["total_daily"].(int64)

		// 格式化流量显示
		upFormatted := formatBytes(dailyUp)
		downFormatted := formatBytes(dailyDown)
		totalFormatted := formatBytes(totalDaily)

		csvContent += fmt.Sprintf("%s,%d,%d,%d,%s,%s,%s\n",
			date, dailyUp, dailyDown, totalDaily, upFormatted, downFormatted, totalFormatted)
	}

	// 设置响应头
	filename := fmt.Sprintf("端口历史数据_%s_%s_%s.csv", serviceName, tag, time.Now().In(time.Local).Format("20060102"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(csvContent)))

	c.String(http.StatusOK, csvContent)
}

// 下载用户历史数据
func (api *DatabaseAPI) DownloadUserHistory(c *gin.Context) {
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
	var ipAddress, serviceName, userEmail string

	userQuery := `
		SELECT 
			s.ip_address,
			s.service_name,
			ct.email
		FROM client_traffics ct
		JOIN services s ON ct.service_id = s.id
		WHERE ct.service_id = ? AND ct.email = ?
	`

	serviceIDInt, err := strconv.Atoi(serviceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "无效的服务ID",
		})
		return
	}

	err = api.db.db.QueryRow(userQuery, serviceIDInt, email).Scan(
		&ipAddress, &serviceName, &userEmail,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "用户信息不存在: " + err.Error(),
		})
		return
	}

	// 获取所有历史流量数据
	historyQuery := `
		SELECT 
			cth.date,
			cth.daily_up,
			cth.daily_down,
			cth.daily_up + cth.daily_down as total_daily
		FROM client_traffic_history cth
		WHERE cth.service_id = ? AND cth.email = ?
		ORDER BY cth.date DESC
	`

	rows, err := api.db.db.Query(historyQuery, serviceIDInt, email)
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

	// 生成CSV内容
	csvContent := "日期,上传流量(Bytes),下载流量(Bytes),总流量(Bytes),上传流量(格式化),下载流量(格式化),总流量(格式化)\n"

	for _, record := range history {
		date := record["date"].(string)
		dailyUp := record["daily_up"].(int64)
		dailyDown := record["daily_down"].(int64)
		totalDaily := record["total_daily"].(int64)

		// 格式化流量显示
		upFormatted := formatBytes(dailyUp)
		downFormatted := formatBytes(dailyDown)
		totalFormatted := formatBytes(totalDaily)

		csvContent += fmt.Sprintf("%s,%d,%d,%d,%s,%s,%s\n",
			date, dailyUp, dailyDown, totalDaily, upFormatted, downFormatted, totalFormatted)
	}

	// 设置响应头
	filename := fmt.Sprintf("用户历史数据_%s_%s_%s.csv", serviceName, email, time.Now().In(time.Local).Format("20060102"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(csvContent)))

	c.String(http.StatusOK, csvContent)
}

// 格式化字节数
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
