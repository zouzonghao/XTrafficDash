package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"xtrafficdash/database"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 配置结构体
type Config struct {
	ListenPort   int    `json:"listen_port"`
	DebugMode    bool   `json:"debug_mode"`
	LogLevel     string `json:"log_level"`
	DatabasePath string `json:"database_path"`
}

// 响应数据结构体
type ResponseData struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

var (
	config *Config
	logger *logrus.Logger
	db     *database.Database
)

// 环境变量读取函数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

func setTimezone() {
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "Asia/Shanghai"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		panic("无效的时区: " + tz)
	}
	time.Local = loc
}

func init() {
	setTimezone()
	// 初始化日志
	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	// 从环境变量读取配置
	config = &Config{
		ListenPort:   getEnvAsInt("LISTEN_PORT", 37022),
		DebugMode:    getEnvAsBool("DEBUG_MODE", false),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		DatabasePath: getEnv("DATABASE_PATH", "xtrafficdash.db"),
	}

	// 设置日志级别
	switch config.LogLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// 初始化JWT
	database.InitJWT()

	// 初始化数据库
	var err error
	db, err = database.OpenDatabase(config.DatabasePath)
	if err != nil {
		logger.Errorf("初始化数据库失败: %v", err)
	} else {
		logger.Info("数据库初始化成功")
	}

	// 初始化hy2配置表
	if db != nil {
		err := db.InitHy2ConfigTable()
		if err != nil {
			logger.Errorf("初始化hy2配置表失败: %v", err)
		}
	}
}

func main() {
	logger.Info("启动XTrafficDash...")
	logger.Infof("监听端口: %d", config.ListenPort)
	logger.Infof("数据库路径: %s", config.DatabasePath)

	// 设置Gin模式
	if config.DebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin路由
	r := gin.New()

	// 使用中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// 设置路由
	setupRoutes(r)

	// 启动服务器
	addr := fmt.Sprintf("0.0.0.0:%d", config.ListenPort)
	logger.Infof("服务器启动在地址 %s", addr)

	// 启动hy2流量同步定时任务（自动执行）
	go startHy2SyncTask()

	if err := r.Run(addr); err != nil {
		logger.Fatalf("服务器启动失败: %v", err)
	}
}

// 设置路由
func setupRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", healthCheck)
	r.GET("/api/health", healthCheck)

	// 认证相关路由（不需要认证）
	r.POST("/api/auth/login", database.HandleLogin)
	r.GET("/api/auth/verify", database.HandleVerifyToken)

	// 接收流量数据的API接口
	r.POST("/api/traffic", handleTraffic)

	// 注册数据库API路由（需要认证）
	if db != nil {
		dbAPI := database.NewDatabaseAPI(db)
		dbAPI.RegisterRoutes(r)
	}

	// 静态文件服务（用于前端）
	// 尝试多个可能的路径
	webDistPaths := []string{
		"../web/dist",   // 开发环境（从backend目录运行）
		"./web/dist",    // 开发环境（从项目根目录运行）
		"/app/web/dist", // Docker环境
	}

	var webDistPath string
	for _, path := range webDistPaths {
		if _, err := os.Stat(path); err == nil {
			webDistPath = path
			logger.Infof("找到web/dist目录: %s", path)
			break
		}
	}

	if webDistPath != "" {
		r.Static("/assets", webDistPath+"/assets")
		r.StaticFile("/", webDistPath+"/index.html")
		r.StaticFile("/favicon.svg", webDistPath+"/favicon.svg")
		r.StaticFile("/site.webmanifest", webDistPath+"/site.webmanifest")
	} else {
		logger.Warn("未找到web/dist目录，静态文件服务将不可用")
	}

	// 添加调试路由
	r.GET("/debug", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Debug endpoint working",
			"time":    time.Now(),
			"path":    c.Request.URL.Path,
		})
	})

	r.GET("/api/hy2-configs", getAllHy2ConfigsHandler)
	r.POST("/api/hy2-configs", saveAllHy2ConfigsHandler)
	r.POST("/api/hy2-configs/add", addHy2ConfigHandler)
	r.POST("/api/hy2-configs/update", updateHy2ConfigHandler)
	r.DELETE("/api/hy2-configs/:id", deleteHy2ConfigHandler)

	// 处理所有其他静态文件请求
	r.NoRoute(func(c *gin.Context) {
		logger.Infof("NoRoute: %s", c.Request.URL.Path)
		// 如果不是API请求，返回index.html（用于SPA路由）
		if !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.File("/app/web/dist/index.html")
		} else {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
		}
	})
}

// CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// 健康检查
func healthCheck(c *gin.Context) {
	dbStatus := "disconnected"
	if db != nil {
		dbStatus = "connected"
	}

	c.JSON(200, ResponseData{
		Success: true,
		Message: "服务正常运行",
		Data: map[string]interface{}{
			"timestamp": time.Now(),
			"version":   "2.0.0",
			"status":    "healthy",
			"database":  dbStatus,
		},
	})
}

// 处理流量数据的专用处理器
func handleTraffic(c *gin.Context) {
	// 读取请求体
	bodyBytes, err := c.GetRawData()
	if err != nil {
		logger.Errorf("读取请求体失败: %v", err)
		c.JSON(400, ResponseData{
			Success: false,
			Error:   "读取请求体失败",
		})
		return
	}

	// 优先读取 X-Real-Ip header
	realIP := c.GetHeader("X-Real-Ip")
	if realIP == "" {
		realIP = c.ClientIP()
	}

	// 构建请求数据
	requestData := map[string]interface{}{
		"timestamp":    time.Now(),
		"method":       c.Request.Method,
		"path":         c.Request.URL.Path,
		"headers":      c.Request.Header,
		"query_params": c.Request.URL.Query(),
		"raw_body":     string(bodyBytes),
		"client_ip":    realIP,
		"user_agent":   c.Request.UserAgent(),
	}

	// 简化日志输出
	logger.Infof("收到流量数据请求 - IP: %s, 数据长度: %d bytes", requestData["client_ip"], len(requestData["raw_body"].(string)))

	// 处理数据库存储
	if db != nil {
		// 尝试解析为流量数据
		var trafficData database.TrafficData
		if err := json.Unmarshal(bodyBytes, &trafficData); err == nil {
			// 成功解析为流量数据，存储到数据库
			err = db.ProcessTrafficData(requestData["client_ip"].(string), requestData["user_agent"].(string), requestData["raw_body"].(string), &trafficData)
			if err != nil {
				logger.Errorf("存储流量数据失败: %v", err)
			} else {
				logger.Infof("流量数据已存储到数据库")
			}
		} else {
			logger.Warnf("请求体不是有效的流量数据格式: %v", err)
		}
	}

	c.JSON(200, ResponseData{
		Success: true,
		Message: "流量数据接收成功",
		Data: map[string]interface{}{
			"timestamp": requestData["timestamp"],
		},
	})
}

// 获取hy2配置
func getHy2ConfigHandler(c *gin.Context) {
	if db == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未初始化"})
		return
	}
	cfgs, err := db.GetAllHy2Configs()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}
	if len(cfgs) == 0 {
		c.JSON(404, gin.H{"success": false, "error": "未找到hy2配置"})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": cfgs[0]}) // 假设只有一个hy2配置
}

// 更新hy2配置
func updateHy2ConfigHandler(c *gin.Context) {
	if db == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未初始化"})
		return
	}
	var cfg database.Hy2Config
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误: " + err.Error()})
		return
	}
	err := db.UpdateHy2Config(&cfg)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "message": "保存成功"})
}

// 删除单条
func deleteHy2ConfigHandler(c *gin.Context) {
	if db == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未初始化"})
		return
	}
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误: id无效"})
		return
	}
	err = db.DeleteHy2Config(id)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}

// 获取全部hy2配置
func getAllHy2ConfigsHandler(c *gin.Context) {
	if db == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未初始化"})
		return
	}
	cfgs, err := db.GetAllHy2Configs()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": cfgs})
}

func isValidHost(host string) bool {
	if host == "" {
		return false
	}
	// 简单IP或域名校验
	ipRe := regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}$`)
	domainRe := regexp.MustCompile(`^([a-zA-Z0-9\-]+\.)+[a-zA-Z]{2,}$`)
	return ipRe.MatchString(host) || domainRe.MatchString(host)
}

func isValidPort(port string) bool {
	p, err := strconv.Atoi(port)
	return err == nil && p > 0 && p <= 65535
}

func isValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// 批量保存（全量覆盖）
func saveAllHy2ConfigsHandler(c *gin.Context) {
	if db == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未初始化"})
		return
	}
	var cfgs []database.Hy2Config
	if err := c.ShouldBindJSON(&cfgs); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误: " + err.Error()})
		return
	}
	// 校验
	if len(cfgs) > 0 {
		// 统一目标地址
		targetURL := cfgs[0].TargetAPIURL
		if !isValidURL(targetURL) {
			c.JSON(400, gin.H{"success": false, "error": "目标API地址无效，必须以http://或https://开头"})
			return
		}
		for i, cfg := range cfgs {
			if !isValidHost(cfg.SourceAPIHost) {
				c.JSON(400, gin.H{"success": false, "error": "第" + strconv.Itoa(i+1) + "行：hy2服务端IP/域名无效"})
				return
			}
			if !isValidPort(cfg.SourceAPIPort) {
				c.JSON(400, gin.H{"success": false, "error": "第" + strconv.Itoa(i+1) + "行：hy2服务端端口无效"})
				return
			}
			if strings.TrimSpace(cfg.SourceAPIPassword) == "" {
				c.JSON(400, gin.H{"success": false, "error": "第" + strconv.Itoa(i+1) + "行：hy2服务端密码不能为空"})
				return
			}
			if cfg.TargetAPIURL != targetURL {
				c.JSON(400, gin.H{"success": false, "error": "所有配置的目标API地址必须一致"})
				return
			}
		}
	}
	// 先清空表再插入
	err := db.DeleteAllHy2Configs()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}
	for _, cfg := range cfgs {
		db.AddHy2Config(&cfg)
	}
	c.JSON(200, gin.H{"success": true, "message": "保存成功"})
}

// 新增单条
func addHy2ConfigHandler(c *gin.Context) {
	if db == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未初始化"})
		return
	}
	var cfg database.Hy2Config
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误: " + err.Error()})
		return
	}
	err := db.AddHy2Config(&cfg)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "message": "添加成功"})
}

// hy2流量同步定时任务（自动执行，支持多配置）
func startHy2SyncTask() {
	for {
		if db == nil {
			time.Sleep(10 * time.Second)
			continue
		}
		cfgs, err := db.GetAllHy2Configs()
		if err != nil {
			logger.Errorf("读取hy2配置失败: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// 如果没有配置，跳过本次执行
		if len(cfgs) == 0 {
			time.Sleep(10 * time.Second)
			continue
		}

		// 从第一个配置中获取目标地址，所有配置共享同一个目标
		targetURL := cfgs[0].TargetAPIURL
		if targetURL == "" {
			logger.Warnf("[HY2] 目标地址为空，跳过本次同步")
			time.Sleep(10 * time.Second)
			continue
		}

		// 为每个配置执行同步，但都发送到同一个目标地址
		for _, cfg := range cfgs {
			// 跳过无效配置
			if cfg.SourceAPIHost == "" || cfg.SourceAPIPort == "" || cfg.SourceAPIPassword == "" {
				continue
			}

			// 创建配置副本，使用统一的目标地址
			syncCfg := database.Hy2Config{
				ID:                cfg.ID,
				SourceAPIPassword: cfg.SourceAPIPassword,
				SourceAPIHost:     cfg.SourceAPIHost,
				SourceAPIPort:     cfg.SourceAPIPort,
				TargetAPIURL:      targetURL, // 使用统一的目标地址
			}

			go hy2SyncOnce(&syncCfg)
		}
		time.Sleep(10 * time.Second)
	}
}

// hy2流量同步单次执行逻辑
func hy2SyncOnce(cfg *database.Hy2Config) {
	client := &http.Client{Timeout: 15 * time.Second}
	// 构建源API URL
	sourceURL := "http://" + cfg.SourceAPIHost + ":" + cfg.SourceAPIPort + "/traffic?clear=1"
	// 1. 拉取源API流量
	req, err := http.NewRequest("GET", sourceURL, nil)
	if err != nil {
		logger.Errorf("[HY2] 创建请求失败: %v", err)
		return
	}
	req.Header.Set("Authorization", cfg.SourceAPIPassword)
	resp, err := client.Do(req)
	if err != nil {
		logger.Errorf("[HY2] 请求源API失败: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		logger.Errorf("[HY2] 源API返回状态码: %d, 响应: %s", resp.StatusCode, string(body))
		return
	}

	var raw struct {
		User struct {
			Tx int64 `json:"tx"`
			Rx int64 `json:"rx"`
		} `json:"user"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("[HY2] 读取源API响应失败: %v", err)
		return
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		logger.Errorf("[HY2] 解析源API响应失败: %v", err)
		return
	}
	logger.Infof("[HY2] 获取到流量数据: tx=%d, rx=%d", raw.User.Tx, raw.User.Rx)

	// 2. 转换格式
	postData := map[string]interface{}{
		"inboundTraffics": []map[string]interface{}{
			{
				"IsInbound":  true,
				"IsOutbound": false,
				"Tag":        "hysteria2",
				"Up":         raw.User.Tx,
				"Down":       raw.User.Rx,
			},
		},
	}
	jsonBytes, _ := json.Marshal(postData)

	// 3. POST到目标API
	postReq, err := http.NewRequest("POST", cfg.TargetAPIURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		logger.Errorf("[HY2] 创建POST请求失败: %v", err)
		return
	}
	postReq.Header.Set("Content-Type", "application/json")
	// 新增：带上真实IP
	postReq.Header.Set("X-Real-Ip", cfg.SourceAPIHost)
	postResp, err := client.Do(postReq)
	if err != nil {
		logger.Errorf("[HY2] 发送POST到目标API失败: %v", err)
		return
	}
	defer postResp.Body.Close()
	if postResp.StatusCode != 200 {
		respBody, _ := io.ReadAll(postResp.Body)
		logger.Errorf("[HY2] 目标API返回状态码: %d, 响应: %s", postResp.StatusCode, string(respBody))
		return
	}
	logger.Infof("[HY2] 流量数据已成功推送到目标API")
}
