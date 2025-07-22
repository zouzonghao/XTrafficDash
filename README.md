# XTrafficDash

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![Vue Version](https://img.shields.io/badge/Vue-3.0+-green.svg)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://www.docker.com/)

一个现代化的3X-UI流量统计面板，使用Vue3 + Go构建，支持多服务器流量监控和可视化。





## 🚀 快速开始

### docker run

```sh
docker run -d \
  --name xtrafficdash \
  -p 37022:37022 \
  -e DATABASE_PATH=/app/data/xtrafficdash.db \
  -e PASSWORD=admin123 \
  -e TZ=Asia/Shanghai \
  --log-opt max-size=5m \
  --log-opt max-file=3 \
  --restart unless-stopped \
  sanqi37/xtrafficdash
```
### docker compose 部署

```
version: '3.8'

services:
  xtrafficdash:
    image: sanqi37/xtrafficdash 
    container_name: xtrafficdash
    restart: unless-stopped
    ports:
      - "37022:37022"
    environment:
      - TZ=Asia/Shanghai
      - DATABASE_PATH=/app/data/xtrafficdash.db
      - PASSWORD=admin123
    logging:
      options:
        max-size: "5m"
        max-file: "3"
```

- 修改 `PASSWORD` ，前端 web 密码，不修改则默认为 admin123

###  3x-ui 接入（需要较新版本）
-  -> 面板设置 
-   -> 常规 
-   -> 外部流量 
- -> 外部流量通知URL 
- -> `http://111.111.111.111:37022/api/traffic`

- 改为自己服务器地址


### hysteria2 接入

#### 1. 修改配置文件
```sh
nano /etc/hysteria/config.yaml
```
##### 添加（请修改 passwd）

```yml
trafficStats:
  listen: :37023 
  secret: passwd 
```
#### 2. 在首页点击 `HY2设置` 进行添加



## 🚀 更新
```bash
# 1. 停止正在运行的容器，防止数据库写入冲突
docker stop xtrafficdash

# 2. 从容器中导出当前数据库文件到宿主机指定目录（备份）
mkdir /usr/xtrafficdash/ 

docker cp xtrafficdash:/app/data /usr/xtrafficdash/

# 3. 修改数据库文件权限，确保后续容器可读写
chmod -R 666 /usr/xtrafficdash

# 4. 删除旧容器（不会影响备份的数据库文件）
docker rm xtrafficdash  

# 5. 删除旧镜像（可选，确保拉取最新镜像）
docker rmi sanqi37/xtrafficdash  

# 6. 重新运行新容器，挂载数据库文件，并设置日志轮转
docker run -d \
  --name xtrafficdash \
  -p 37022:37022 \
  -e DATABASE_PATH=/app/data/xtrafficdash.db \
  -e TZ=Asia/Shanghai \
  -e PASSWORD=admin123 \
  --log-opt max-size=5m \
  --log-opt max-file=3 \
  --restart unless-stopped \
  sanqi37/xtrafficdash

# 7. 再次停止新容器，导入备份的数据库文件
docker stop xtrafficdash

# 8. 将备份的数据库文件拷贝回新容器
docker cp /usr/xtrafficdash/data xtrafficdash:/app/data

# 9. 启动新容器，使用导入的数据库
docker start xtrafficdash
```

## 🛠️ 技术栈

### 后端
- **Go 1.21+**: 高性能后端服务
- **Gin**: Web框架，支持中间件和路由
- **SQLite**: 轻量级数据库，支持连接池优化
- **JWT**: 身份认证和会话管理
- **Logrus**: 结构化日志记录

### 前端
- **Vue 3**: 渐进式JavaScript框架，使用Composition API
- **Vite 6.x**: 快速构建工具，支持热重载
- **Pinia**: 状态管理，替代Vuex
- **Vue Router**: 客户端路由
- **Chart.js**: 数据可视化图表库
- **Axios**: HTTP客户端
- **Tailwind CSS**: 原子化CSS框架

## 🏗️ 项目结构

```
x-ui-panel/
│
├── backend/ # Go 后端服务
│ ├── main.go # 后端主入口，静态文件服务
│ ├── database/ # 数据库相关
│ │ ├── api.go # API 处理
│ │ ├── auth.go # JWT 认证
│ │ └── database.go # 数据库连接与操作
│ ├── go.mod # Go 依赖管理
│ └── go.sum # Go 依赖锁定
│
├── web/ # Vue3 前端
│ ├── src/
│ │ ├── assets/ # 静态资源与样式
│ │ │ └── main.css
│ │ ├── components/ # 通用组件
│ │ │ ├── EditNameModal.vue   # 改名
│ │ │ └── ServiceCard.vue   # 首页卡片
│ │ ├── router/ # 路由配置
│ │ │ └── index.js
│ │ ├── stores/ # Pinia 状态管理
│ │ │ ├── auth.js
│ │ │ └── services.js
│ │ ├── utils/ # 工具函数
│ │ │ ├── api.js
│ │ │ └── formatters.js
│ │ ├── views/ # 页面组件
│ │ │ ├── Detail.vue
│ │ │ ├── Home.vue
│ │ │ ├── Hy2Setting.vue
│ │ │ ├── Login.vue
│ │ │ ├── PortDetail.vue
│ │ │ └── UserDetail.vue
│ │ ├── App.vue # 根组件
│ │ └── main.js # 前端入口
│ ├── public/ # 公共静态资源
│ │ ├── favicon.svg
│ │ └── site.webmanifest
│ ├── index.html # HTML 入口
│ ├── package.json # 前端依赖
│ ├── package-lock.json # 依赖锁定
│ └── vite.config.js # Vite 配置
│
├── Dockerfile # Docker 构建文件
├── docker-compose.yml # Docker Compose 根配置
└── README.md # 项目说明文档
```

## 🛠️ 开发相关

### Docker自己编译部署

```bash
# 1. 克隆项目
git clone <repository-url>
cd xtrafficdash

# 2. 启动服务
docker-compose up -d

# 3. 访问面板
# 打开浏览器访问: http://localhost:37022
# 默认密码: admin123
```

### 本地开发

```bash
# 1. 启动后端
cd backend
export PASSWORD=admin123
go run main.go

# 2. 启动前端（新终端）
cd web
npm install
npm run dev

# 3. 访问前端
# 打开浏览器访问: http://localhost:3000
# 后端API地址: http://localhost:37022
```

### 编译可执行文件

```bash
# 编译后端服务
cd backend
go build -o main main.go

# 或者指定输出文件名
go build -o xtrafficdash main.go

# 运行编译后的程序
./main
# 或
./xtrafficdash
```

## 🔧 配置说明

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PASSWORD` | `admin123` | 登录密码（必填） |
| `TZ` | `Asia/Shanghai` | 容器/服务时区设置 |
| `LISTEN_PORT` | `37022` | 服务监听端口 |
| `DEBUG_MODE` | `true` | 调试模式 |
| `LOG_LEVEL` | `info` | 日志级别 |
| `DATABASE_PATH` | `xtrafficdash.db` | 数据库文件路径 |

### 静态文件服务

后端支持智能路径检测，自动适配不同部署环境：
- **开发环境**: 从backend目录运行时使用 `../web/dist`
- **项目根目录**: 从项目根目录运行时使用 `./web/dist`
- **Docker环境**: 容器内使用 `/app/web/dist`


## 🔒 安全说明

- 所有API接口（除登录外）都需要JWT认证
- 密码通过环境变量配置，支持Docker部署
- 支持CORS跨域配置
- 数据库使用SQLite，数据文件可持久化


### 日志查看
```bash
# 查看Docker容器日志
docker-compose logs -f

# 查看前端构建日志
cd web && npm run build

# 查看后端启动日志
cd backend && go run main.go

# 测试静态文件服务
curl http://localhost:37022/favicon.svg
curl http://localhost:37022/site.webmanifest
```

## 📄 许可证

本项目采用 MIT 许可证。

