package database

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginRequest struct {
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

var jwtSecret []byte

// 生成带16位固定后缀的JWT密钥
func getPasswordWithSuffix() string {
	password := os.Getenv("PASSWORD")
	if password == "" {
		return ""
	}
	h := md5.New()
	h.Write([]byte(password))
	hash := hex.EncodeToString(h.Sum(nil))
	return password + hash[:16]
}

// 初始化JWT密钥
func InitJWT() {
	// 从环境变量读取密钥并加上16位固定后缀
	secret := getPasswordWithSuffix()
	if secret == "" {
		panic("必须设置环境变量 PASSWORD 作为 JWT 密钥和登录密码")
	}
	jwtSecret = []byte(secret)
}

// 验证密码
func validatePassword(password string) bool {
	// 从环境变量读取密码
	envPassword := os.Getenv("PASSWORD")
	if envPassword == "" {
		// 未设置密码，直接返回 false
		return false
	}
	return password == envPassword
}

// 生成JWT token
func generateToken() (string, error) {
	claims := Claims{
		UserID: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// 验证JWT token
func validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// 登录处理
func HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{
			Success: false,
			Message: "请求参数错误",
		})
		return
	}

	if !validatePassword(req.Password) {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Message: "密码错误",
		})
		return
	}

	token, err := generateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, LoginResponse{
			Success: false,
			Message: "生成token失败",
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Success: true,
		Message: "登录成功",
		Token:   token,
	})
}

// 验证token中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "未提供认证token"})
			c.Abort()
			return
		}

		// 提取Bearer token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "token格式错误"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]
		claims, err := validateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "token无效"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// 验证token接口
func HandleVerifyToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "token有效",
	})
}
