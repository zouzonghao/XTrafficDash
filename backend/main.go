package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"api-proxy-service/database"

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

func init() {
	// 初始化日志
	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	// 从环境变量读取配置
	config = &Config{
		ListenPort:   getEnvAsInt("LISTEN_PORT", 37022),
		DebugMode:    getEnvAsBool("DEBUG_MODE", false),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		DatabasePath: getEnv("DATABASE_PATH", "xui_traffic.db"),
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
}

func main() {
	logger.Info("启动X-UI流量统计面板...")
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
	addr := fmt.Sprintf(":%d", config.ListenPort)
	logger.Infof("服务器启动在端口 %d", config.ListenPort)

	if err := r.Run(addr); err != nil {
		logger.Fatalf("服务器启动失败: %v", err)
	}
}

// 设置路由
func setupRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", healthCheck)

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
	r.Static("/static", "./web/dist")
	r.StaticFile("/", "./web/dist/index.html")
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

	// 构建请求数据
	requestData := map[string]interface{}{
		"timestamp":    time.Now(),
		"method":       c.Request.Method,
		"path":         c.Request.URL.Path,
		"headers":      c.Request.Header,
		"query_params": c.Request.URL.Query(),
		"raw_body":     string(bodyBytes),
		"client_ip":    c.ClientIP(),
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
