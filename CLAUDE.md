# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Personal blog system built with Go (Gin) backend, vanilla JS frontend, MySQL database, and Docker Compose deployment with Nginx reverse proxy.

## Commands

### Docker Deployment
```bash
# Build and start all services
DOCKER_BUILDKIT=0 docker compose up -d --build

# View logs
docker compose logs -f backend

# Stop services
docker compose down
```

### Local Go Development
```bash
cd backend
go mod tidy
go run main.go  # Requires local MySQL instance
```

### Database Operations
```bash
# Access MySQL shell
docker compose exec mysql mysql -uroot -p

# Backup database
docker compose exec mysql mysqldump -uroot -p blog > backup.sql
```

## Architecture

### Backend (Go/Gin)
- **Entry point**: `backend/main.go` - Connects to DB, runs auto-migration, creates default admin/category, starts server
- **Routes**: `backend/routes/routes.go` - All API endpoints organized into public, auth, and admin groups
- **Controllers**: `backend/controllers/` - Handle HTTP requests for articles, categories, comments, auth, RSS
- **Models**: `backend/models/models.go` - GORM models: User, Article, Category, Comment
- **Middleware**: `backend/middleware/auth.go` - Session-based authentication for admin routes
- **Config**: `backend/config/config.go` - Environment variable loading

### Frontend (Static)
- **Pages**: `frontend/*.html` - index, article, admin, login, about
- **JS**: `frontend/static/js/app.js` (public), `frontend/static/js/admin.js` (admin panel)
- **CSS**: `frontend/static/css/style.css` - CSS variables for theming, supports dark mode

### Database Models
- **User**: id, username, password (bcrypt), email
- **Article**: id, title, slug, content (HTML from Quill), summary, cover_image, category_id, tags, view_count, is_published
- **Category**: id, name, slug
- **Comment**: id, article_id, nickname, email, content, is_approved

### API Structure
- `/api/articles`, `/api/categories`, `/api/comments`, `/api/rss` - Public endpoints
- `/api/auth/login`, `/api/auth/logout`, `/api/auth/me` - Authentication
- `/api/admin/*` - Protected admin endpoints (require session authentication)

### Service Stack (Docker Compose)
- **nginx**: Serves static files, proxies `/api/*` to backend
- **backend**: Go application on port 8080
- **mysql**: MySQL 8.0 with health check, data persisted to `./data/mysql`

## Environment Variables (.env)
- `DB_PASSWORD`, `DB_USER`, `DB_NAME` - Database config
- `ADMIN_USERNAME`, `ADMIN_PASSWORD` - Default admin credentials
- `SESSION_SECRET` - Cookie session secret

## Default Credentials
- Username: `admin`, Password: `admin123` (change after first login)