# My Blog - 个人博客系统

一个现代化的个人博客系统，使用 Go + MySQL + Nginx 构建，支持 Docker Compose 一键部署。

## 功能特性

### 前台功能
- 📝 文章列表与详情页
- 🔍 文章搜索
- 📂 分类筛选
- 💬 评论系统
- 🌙 暗色模式
- 📱 响应式设计
- 🔖 RSS 订阅
- 🎨 代码高亮

### 后台功能
- ✏️ 富文本编辑器 (Quill.js)
- 📊 文章管理 (增删改查)
- 📁 分类管理
- 💭 评论审核
- 🔐 用户认证

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端框架 | Go + Gin |
| ORM | GORM |
| 数据库 | MySQL 8.0 |
| 前端 | HTML5 + CSS3 + JavaScript |
| 编辑器 | Quill.js |
| 代码高亮 | Highlight.js |
| 容器化 | Docker + Docker Compose |
| 反向代理 | Nginx |

## 快速开始

### 前置要求

- Docker
- Docker Compose

### 部署步骤

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd zxpblog
   ```

2. **修改配置**

   编辑 `.env` 文件，修改以下配置：
   ```env
   # 数据库密码（请修改为强密码）
   DB_PASSWORD=your_secure_password_here

   # 管理员账号（首次启动后请及时修改密码）
   ADMIN_USERNAME=admin
   ADMIN_PASSWORD=admin123

   # Session 密钥（请修改为随机字符串）
   SESSION_SECRET=your_session_secret_key_here
   ```

3. **启动服务**
   ```bash
   DOCKER_BUILDKIT=0 docker compose up -d --build
   ```

4. **访问博客**
   - 前台首页: http://localhost
   - 后台管理: http://localhost/admin.html
   - 登录页面: http://localhost/login.html

### 默认账号

- 用户名: `admin`
- 密码: `admin123`

> ⚠️ **重要**: 首次登录后请立即修改默认密码！

## 目录结构

```
zxpblog/
├── backend/                 # Go 后端代码
│   ├── main.go             # 入口文件
│   ├── config/             # 配置管理
│   ├── models/             # 数据模型
│   ├── controllers/        # 控制器
│   ├── routes/             # 路由
│   ├── middleware/         # 中间件
│   ├── go.mod
│   └── Dockerfile
├── frontend/               # 前端静态文件
│   ├── index.html          # 首页
│   ├── article.html        # 文章详情
│   ├── admin.html          # 后台管理
│   ├── login.html          # 登录页面
│   ├── about.html          # 关于页面
│   └── static/
│       ├── css/style.css   # 样式文件
│       └── js/
│           ├── app.js      # 前端脚本
│           └── admin.js    # 后台脚本
├── nginx/
│   └── nginx.conf          # Nginx 配置
├── data/                    # 数据持久化目录
│   └── mysql/              # MySQL 数据
├── docker-compose.yml      # Docker Compose 配置
├── .env                    # 环境变量
└── README.md
```

## API 接口

### 公开接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/articles` | 获取文章列表 |
| GET | `/api/articles/:id` | 获取文章详情 |
| GET | `/api/categories` | 获取分类列表 |
| GET | `/api/comments` | 获取评论列表 |
| POST | `/api/comments` | 提交评论 |
| GET | `/api/rss` | RSS 订阅 |

### 认证接口

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | `/api/auth/login` | 登录 |
| POST | `/api/auth/logout` | 登出 |
| GET | `/api/auth/me` | 获取当前用户 |

### 管理接口 (需要认证)

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/admin/articles` | 获取所有文章 |
| POST | `/api/admin/articles` | 创建文章 |
| PUT | `/api/admin/articles/:id` | 更新文章 |
| DELETE | `/api/admin/articles/:id` | 删除文章 |
| POST | `/api/admin/categories` | 创建分类 |
| PUT | `/api/admin/categories/:id` | 更新分类 |
| DELETE | `/api/admin/categories/:id` | 删除分类 |
| GET | `/api/admin/comments` | 获取所有评论 |
| PUT | `/api/admin/comments/:id/approve` | 审核评论 |
| DELETE | `/api/admin/comments/:id` | 删除评论 |

## 数据持久化

数据存储在 `./data/mysql` 目录下，即使容器重启也不会丢失数据。

## 常用命令

```bash
# 启动服务
docker compose up -d

# 停止服务
docker compose down

# 重新构建并启动
DOCKER_BUILDKIT=0 docker compose up -d --build

# 查看日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f backend

# 进入 MySQL 容器
docker compose exec mysql mysql -uroot -p

# 备份数据库
docker compose exec mysql mysqldump -uroot -p blog > backup.sql
```

## 自定义配置

### 修改端口

编辑 `docker-compose.yml`，修改 nginx 服务的端口映射：
```yaml
nginx:
  ports:
    - "8080:80"  # 将 80 改为其他端口
```

### 配置域名

编辑 `nginx/nginx.conf`，修改 `server_name`：
```nginx
server {
    listen 80;
    server_name your-domain.com;
    # ...
}
```

### HTTPS 配置

推荐使用 Certbot 配置 SSL 证书，或修改 nginx.conf 添加 SSL 配置。

## 故障排除

### 容器无法启动

1. 检查端口是否被占用
2. 查看日志: `docker compose logs`
3. 确保 `.env` 文件配置正确

### 数据库连接失败

1. 等待 MySQL 完全启动（健康检查通过）
2. 检查数据库密码是否正确
3. 查看 MySQL 日志: `docker compose logs mysql`

### 无法登录

1. 确认使用正确的用户名和密码
2. 清除浏览器 Cookie 后重试
3. 检查 SESSION_SECRET 配置

## 开发

### 本地开发

```bash
# 进入后端目录
cd backend

# 下载依赖
go mod tidy

# 运行（需要本地 MySQL）
go run main.go
```

## License

MIT License