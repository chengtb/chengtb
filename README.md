# 餐厅扫码点餐系统

A full-stack restaurant QR code ordering system with real-time multi-user cart synchronization, intelligent cooking task dispatch, and three independent frontend applications.

## Tech Stack

- **Backend**: Golang + Gin (RESTful API + WebSocket)
- **Database**: MySQL 8
- **Customer App**: Vue3 + Vant 4 (mobile H5)
- **Merchant App**: Vue3 + Vuetify 3 (PC dashboard)
- **Chef App**: Vue3 + Vuetify 3 (responsive)

## Project Structure

```
├── backend/          # Go/Gin backend service
├── customer-app/     # Customer ordering app (mobile H5)
├── merchant-app/     # Merchant management dashboard (PC)
├── chef-app/         # Chef task management app
└── docker-compose.yml
```

## Core Features

### Customer Ordering (Port 3001)
- QR code scan → shared multi-user cart per table
- Real-time cart sync via WebSocket (multiple phones, one cart)
- Submit order → live status tracking

### Merchant Management (Port 3002)
- Visual table grid with color-coded status
- Order list with payment and discount management
- Auto/manual cooking task dispatch
- Chef, dish, and recipe management
- System configuration

### Chef Task App (Port 3003)
- Per-chef task list (recipe variant per task, e.g. "宫保鸡丁x2")
- Shows merged table numbers when multiple orders combined
- Mark tasks complete → auto-updates order status

### Auto-Dispatch Engine
1. **Portion aggregation**: Groups pending order items by dish within time window (default 5 min)
2. **Recipe matching**: Exact match first → combination split (e.g. 3 portions = x2 + x1)
3. **Chef assignment**: Skill match (chef_recipe table) → load balancing → priority (VIP, time)

## Quick Start

### Docker Compose (recommended)

```bash
docker-compose up -d
```

Services:
| Service | URL |
|---------|-----|
| Customer App | http://localhost:3001 |
| Merchant App | http://localhost:3002 |
| Chef App | http://localhost:3003 |
| Backend API | http://localhost:8080 |
| MySQL | localhost:3306 |

### Local Development

**Backend:**
```bash
cd backend
export DB_HOST=localhost DB_USER=restaurant DB_PASSWORD=restaurant123 DB_NAME=restaurant
go run main.go
```

**Customer App:**
```bash
cd customer-app
cp .env.example .env
npm install && npm run dev
```

**Merchant App:**
```bash
cd merchant-app
cp .env.example .env
npm install && npm run dev
```

**Chef App:**
```bash
cd chef-app
cp .env.example .env
npm install && npm run dev
```

## Database Schema

Key tables:
- `table` — restaurant tables with status
- `dish` — menu items with category
- `recipe` — recipe variants (e.g. "宫保鸡丁x2", portion=2)
- `chef_recipe` — which chef knows which recipe variant
- `cart_session` — active ordering session per table
- `session_item` — items in cart
- `order` / `order_item` — submitted orders
- `chef` — chef profiles with load tracking
- `cooking_task` — dispatched tasks with merged order/table info
- `system_config` — runtime configuration

## API Overview

| Prefix | Target |
|--------|--------|
| `/api/customer/...` | Customer ordering endpoints |
| `/api/merchant/...` | Merchant management endpoints |
| `/api/chef/...` | Chef task endpoints |
| `/ws/cart/:sessionId` | WebSocket for cart sync |

---

Contact: 893400722@qq.com
