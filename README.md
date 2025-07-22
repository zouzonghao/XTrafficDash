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

- 修改 `X_UI_PASSWORD` ，前端 web 密码，不修改则默认为 admin123

###  3x-ui 接入（需要较新版本）
-  -> 面板设置 
-   -> 常规 
-   -> 外部流量 
- -> 外部流量通知URL 
- -> `http://111.111.111.111:37022/api/traffic`

- 改为自己服务器地址


### hysteria2 接入

#### 1. 修改配置文件
   ```
   nano /etc/hysteria/config.yaml
   ```
   添加

## 🚀 更新
```bash
# 1. 停止正在运行的容器，防止数据库写入冲突
docker stop xtrafficdash

# 2. 从容器中导出当前数据库文件到宿主机指定目录（备份）
mkdir /usr/xtrafficdash/ && docker cp xtrafficdash:/app/data /usr/xtrafficdash/

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

# 7. （可选）再次停止新容器，导入备份的数据库文件
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
├── backend/              # Go后端代码
│   ├── main.go          # 主程序入口（包含智能静态文件服务）
│   ├── database/        # 数据库相关代码
│   │   ├── database.go  # 数据库操作和连接池
│   │   ├── api.go      # API处理器
│   │   └── auth.go     # JWT认证相关
│   ├── go.mod          # Go模块文件
│   └── go.sum          # Go依赖锁定文件
├── web/                 # Vue3前端代码
│   ├── src/
│   │   ├── components/  # Vue组件
│   │   │   ├── ServiceCard.vue    # 服务卡片组件
│   │   │   └── EditNameModal.vue  # 编辑名称模态框
│   │   ├── views/      # 页面组件
│   │   │   ├── Home.vue           # 首页
│   │   │   ├── Login.vue          # 登录页
│   │   │   ├── Detail.vue         # 详情页
│   │   │   ├── PortDetail.vue     # 端口详情页
│   │   │   └── UserDetail.vue     # 用户详情页
│   │   ├── stores/     # Pinia状态管理
│   │   │   ├── auth.js            # 认证状态
│   │   │   └── services.js        # 服务数据状态
│   │   ├── utils/      # 工具函数
│   │   │   ├── api.js             # API请求封装
│   │   │   └── formatters.js      # 数据格式化
│   │   ├── assets/     # 静态资源
│   │   │   └── main.css           # 主样式文件
│   │   ├── router/     # 路由配置
│   │   │   └── index.js           # 路由定义
│   │   ├── App.vue     # 主应用组件
│   │   └── main.js     # 应用入口
│   ├── public/         # 公共静态文件
│   │   ├── favicon.svg # 站点图标
│   │   └── site.webmanifest # Web应用清单
│   ├── package.json    # 前端依赖配置
│   ├── vite.config.js  # Vite构建配置
│   └── index.html      # HTML入口文件
├── configs/            # 配置文件
│   ├── env.example     # 环境变量示例
│   └── docker-compose.yml # Docker编排配置
├── scripts/            # 部署脚本
├── Dockerfile          # Docker多阶段构建文件
├── docker-compose.yml  # Docker编排文件
├── .gitignore          # Git忽略文件
├── .dockerignore       # Docker忽略文件
└── README.md           # 项目说明文档
```

## 🛠️ 开发相关

### Docker自己编译部署

```bash
# 1. 克隆项目
git clone <repository-url>
cd xtrafficdash

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
| `X_UI_PASSWORD` | `admin123` | 登录密码（必填） |
| `LISTEN_PORT` | `37022` | 服务监听端口 |
| `DEBUG_MODE` | `true` | 调试模式 |
| `LOG_LEVEL` | `info` | 日志级别 |
| `DATABASE_PATH` | `xtrafficdash.db` | 数据库文件路径 |

### 静态文件服务

后端支持智能路径检测，自动适配不同部署环境：
- **开发环境**: 从backend目录运行时使用 `../web/dist`
- **项目根目录**: 从项目根目录运行时使用 `./web/dist`
- **Docker环境**: 容器内使用 `/app/web/dist`

### Docker配置

- **端口映射**: `37022:37022`
- **数据持久化**: `./data:/app/data`
- **健康检查**: 自动检测服务状态
- **自动重启**: 容器异常时自动重启

## 📊 功能特性

- 🔐 **安全认证**: 基于JWT的登录验证系统
- 📊 **实时监控**: 多服务器流量数据实时展示
- 📈 **数据可视化**: 使用Chart.js绘制流量趋势图
- 🧩 **HY2多配置管理**: 支持多组HY2服务端配置，每组独立同步流量数据，统一目标API地址
- 📝 **HY2设置页面**: 前端支持增删改多组HY2配置，表格自适应响应式布局
- 🎨 **现代化UI**: 响应式设计，支持移动端
- 🐳 **容器化部署**: 支持Docker一键部署
- 🔄 **自动刷新**: 数据自动更新，无需手动刷新
- ⏰ **智能状态**: 60秒活跃状态判断，减少误判
- 🌍 **时区支持**: 正确处理多时区时间显示
- 📱 **移动适配**: 完美支持手机和平板设备
- 🎯 **站点图标**: 自定义SVG图标和Web应用清单
- 🚀 **静态文件服务**: 智能路径检测，支持多种部署环境
- 🔧 **安全漏洞修复**: 已修复npm安全漏洞，使用稳定版本

---

## 🏗️ 前端页面结构

```
web/src/views/
├── Home.vue         # 首页
├── Login.vue        # 登录页
├── Detail.vue       # 节点详情页
├── PortDetail.vue   # 端口详情页
├── UserDetail.vue   # 用户详情页
├── Hy2Setting.vue   # HY2设置页（多配置管理、全局目标API地址）
```



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

4. **站点图标不显示**
   - 确认favicon.svg文件存在于web/public/目录
   - 检查后端静态文件服务是否正常启动
   - 查看后端日志中的路径检测信息

5. **Docker构建失败**
   - 确认Dockerfile中的多阶段构建配置正确
   - 检查前端依赖是否完整安装
   - 查看构建日志中的具体错误信息

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

### 部署脚本
```bash
# 使用部署脚本构建和推送Docker镜像
./scripts/docker-deploy.sh

# 启动开发环境
./scripts/start.sh

# 测试Docker构建
./test-docker-build.sh
```

## 📄 许可证

本项目采用 MIT 许可证。

