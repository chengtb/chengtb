# 餐厅扫码点餐系统

全栈餐厅扫码点餐系统，支持多用户实时购物车同步、智能烹饪任务派发，包含三个独立前端应用。

## 技术栈

- **后端**：Golang + Gin（RESTful API + WebSocket）
- **数据库**：MySQL 8
- **顾客端**：Vue3 + Vant 4（移动端 H5）
- **商家端**：Vue3 + Vuetify 3（PC 管理后台）
- **厨师端**：Vue3 + Vuetify 3（自适应布局）

## 项目结构

```
├── backend/          # Go/Gin 后端服务
├── customer-app/     # 顾客点餐应用（移动端 H5）
├── merchant-app/     # 商家管理后台（PC 端）
├── chef-app/         # 厨师任务管理应用
└── docker-compose.yml
```

## 核心功能

### 顾客点餐（端口 3001）
- 扫描二维码 → 同桌多人共享购物车
- 通过 WebSocket 实时同步购物车（多部手机，一个购物车）
- 提交订单 → 实时查看订单状态

### 商家管理（端口 3002）
- 餐桌可视化网格，颜色区分桌态
- 订单列表，支持收款与折扣管理
- 自动/手动派发烹饪任务
- 厨师、菜品及菜谱管理
- 系统参数配置

### 厨师任务端（端口 3003）
- 按厨师展示任务列表（每个任务显示菜谱规格，如"宫保鸡丁x2"）
- 多订单合并时显示合并桌号
- 标记任务完成 → 自动更新订单状态

### 自动派发引擎
1. **份量聚合**：在时间窗口内（默认 5 分钟）按菜品汇总待处理订单项
2. **菜谱匹配**：优先精确匹配 → 组合拆分（如 3 份 = x2 + x1）
3. **厨师分配**：技能匹配（chef_recipe 表）→ 负载均衡 → 优先级（VIP、时间）

## 快速启动

### Docker Compose（推荐）

```bash
docker-compose up -d
```

服务列表：
| 服务 | 访问地址 |
|------|---------|
| 顾客端 | http://localhost:3001 |
| 商家端 | http://localhost:3002 |
| 厨师端 | http://localhost:3003 |
| 后端 API | http://localhost:8080 |
| MySQL | localhost:3306 |

### 本地开发

**后端：**
```bash
cd backend
export DB_HOST=localhost DB_USER=restaurant DB_PASSWORD=restaurant123 DB_NAME=restaurant
go run main.go
```

**顾客端：**
```bash
cd customer-app
cp .env.example .env
npm install && npm run dev
```

**商家端：**
```bash
cd merchant-app
cp .env.example .env
npm install && npm run dev
```

**厨师端：**
```bash
cd chef-app
cp .env.example .env
npm install && npm run dev
```

## 数据库结构

主要数据表：
- `table` — 餐桌信息及桌态
- `dish` — 菜单菜品及分类
- `recipe` — 菜谱规格（如"宫保鸡丁x2"，份量=2）
- `chef_recipe` — 厨师技能与菜谱的对应关系
- `cart_session` — 每张桌的点餐会话
- `session_item` — 购物车中的菜品项
- `order` / `order_item` — 已提交的订单
- `chef` — 厨师信息及负载统计
- `cooking_task` — 已派发的烹饪任务（含合并订单/桌号信息）
- `system_config` — 运行时系统配置

## API 接口概览

| 路由前缀 | 说明 |
|---------|------|
| `/api/customer/...` | 顾客点餐接口 |
| `/api/merchant/...` | 商家管理接口 |
| `/api/chef/...` | 厨师任务接口 |
| `/ws/cart/:sessionId` | 购物车同步 WebSocket |

---

联系方式：893400722@qq.com
