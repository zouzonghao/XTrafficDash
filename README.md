# X-UI 流量统计面板

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![Vue Version](https://img.shields.io/badge/Vue-3.0+-green.svg)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://www.docker.com/)

一个现代化的X-UI流量统计面板，使用Vue3 + Go构建，支持多服务器流量监控和可视化。

## 🛠️ 技术栈

### 后端
- **Go 1.21+**: 高性能后端服务
- **Gin**: Web框架
- **SQLite**: 轻量级数据库
- **JWT**: 身份认证

### 前端
- **Vue 3**: 渐进式JavaScript框架
- **Vite**: 快速构建工具
- **Pinia**: 状态管理
- **Chart.js**: 数据可视化
- **Tailwind CSS**: 样式框架

## 🏗️ 项目结构

```
x-ui-panel/
├── backend/              # Go后端代码
│   ├── main.go          # 主程序入口
│   ├── database/        # 数据库相关代码
│   │   ├── database.go  # 数据库操作
│   │   ├── api.go      # API处理器
│   │   └── auth.go     # 认证相关
│   ├── go.mod          # Go模块文件
│   └── go.sum          # Go依赖锁定文件
├── web/                 # Vue3前端代码
│   ├── src/
│   │   ├── components/  # Vue组件
│   │   ├── views/      # 页面组件
│   │   ├── stores/     # Pinia状态管理
│   │   ├── utils/      # 工具函数
│   │   ├── assets/     # 静态资源
│   │   ├── router/     # 路由配置
│   │   ├── App.vue     # 主应用组件
│   │   └── main.js     # 应用入口
│   ├── package.json    # 前端依赖配置
│   ├── vite.config.js  # Vite构建配置
│   └── index.html      # HTML入口文件
├── configs/            # 配置文件
│   ├── env.example     # 环境变量示例
│   └── docker-compose.yml # Docker编排配置
├── scripts/            # 部署脚本
│   ├── start.sh        # 开发环境启动脚本
│   └── docker-deploy.sh # Docker部署脚本
├── Dockerfile          # Docker构建文件
├── .gitignore          # Git忽略文件
└── README.md           # 项目说明文档
```

## 📋 系统要求

- **Go**: 1.21 或更高版本
- **Node.js**: 18 或更高版本
- **Docker**: 20.10 或更高版本（可选）
- **内存**: 至少 512MB RAM
- **存储**: 至少 100MB 可用空间

## 🚀 快速开始

### Docker部署（推荐）

```bash
# 1. 克隆项目
git clone <repository-url>
cd x-ui-panel

# 2. 配置环境变量
cp configs/env.example configs/.env
# 编辑 configs/.env 文件，设置密码

# 3. 启动服务
docker-compose up -d

# 4. 访问面板
# 打开浏览器访问: http://localhost:37022
# 默认密码: admin123
```

### 本地开发

```bash
# 1. 启动后端
cd backend
export X_UI_PASSWORD=your_password
go run main.go

# 2. 启动前端（新终端）
cd web
npm install
npm run dev

# 3. 访问前端
# 打开浏览器访问: http://localhost:3000
```

## 🔧 配置说明

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `X_UI_PASSWORD` | `admin123` | 登录密码（必填） |
| `LISTEN_PORT` | `37022` | 服务监听端口 |
| `DEBUG_MODE` | `true` | 调试模式 |
| `LOG_LEVEL` | `info` | 日志级别 |
| `DATABASE_PATH` | `xui_traffic.db` | 数据库文件路径 |

### Docker配置

- **端口映射**: `37022:37022`
- **数据持久化**: `./data:/app/data`
- **健康检查**: 自动检测服务状态
- **自动重启**: 容器异常时自动重启

## 📊 功能特性

- 🔐 **安全认证**: 基于JWT的登录验证系统
- 📊 **实时监控**: 多服务器流量数据实时展示
- 📈 **数据可视化**: 使用Chart.js绘制流量趋势图
- 🎨 **现代化UI**: 响应式设计，支持移动端
- 🐳 **容器化部署**: 支持Docker一键部署
- 🔄 **自动刷新**: 数据自动更新，无需手动刷新
- ⏰ **智能状态**: 60秒活跃状态判断，减少误判
- 🌍 **时区支持**: 正确处理多时区时间显示
- 📱 **移动适配**: 完美支持手机和平板设备

## 🛠️ 开发指南

### 添加新组件
1. 在 `web/src/components/` 创建组件文件
2. 使用Composition API编写组件逻辑
3. 在需要的页面中导入并使用

### 添加新API
1. 在 `backend/database/api.go` 中添加新的处理方法
2. 在 `RegisterRoutes` 中注册路由
3. 在前端 `web/src/utils/api.js` 中添加对应的API调用

### 数据库迁移
1. 修改 `backend/database/database.go` 中的 `initDatabase` 函数
2. 添加新的表结构或字段
3. 重启服务，数据库会自动迁移

## 🔌 API文档

### 认证接口
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/logout` - 用户登出

### 服务管理
- `GET /api/db/services` - 获取服务列表
- `GET /api/db/service/:id` - 获取服务详情
- `DELETE /api/db/service/:id` - 删除服务

### 流量数据
- `POST /api/traffic` - 接收流量数据
- `GET /api/db/traffic-summary` - 获取流量汇总
- `GET /api/db/traffic-history` - 获取流量历史
- `GET /api/db/weekly-traffic/:id` - 获取周流量数据

### 端口和用户详情
- `GET /api/db/port-detail/:service_id/:tag` - 获取端口详情
- `GET /api/db/user-detail/:service_id/:email` - 获取用户详情

## 🔒 安全说明

- 所有API接口（除登录外）都需要JWT认证
- 密码通过环境变量配置，支持Docker部署
- 支持CORS跨域配置
- 数据库使用SQLite，数据文件可持久化

## 📈 性能优化

- 前端代码分割和懒加载
- 静态资源压缩和缓存
- 数据库查询优化
- 健康检查和自动恢复

## 🐛 故障排除

### 常见问题

1. **登录失败**
   - 检查环境变量 `X_UI_PASSWORD` 是否正确设置
   - 查看后端日志确认认证逻辑

2. **数据库连接失败**
   - 确认SQLite3已安装
   - 检查数据库文件权限

3. **前端无法连接后端**
   - 确认后端服务已启动
   - 检查Vite代理配置

### 日志查看
```bash
# 查看Docker容器日志
docker-compose logs -f

# 查看前端构建日志
cd web && npm run build
```

### 部署脚本
```bash
# 使用部署脚本构建和推送Docker镜像
./scripts/docker-deploy.sh

# 启动开发环境
./scripts/start.sh
```

## 📄 许可证

本项目采用 MIT 许可证。



