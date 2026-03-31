# 餐厅扫码点餐系统 — 需求规格说明书

> 用途：本文档作为向 AI 编程助手提交开发任务的需求输入，描述系统全部功能、数据结构、接口及业务规则。  
> 联系：893400722@qq.com

---

## 目录

1. [项目背景与目标](#1-项目背景与目标)
2. [系统角色](#2-系统角色)
3. [技术选型要求](#3-技术选型要求)
4. [整体架构要求](#4-整体架构要求)
5. [数据库设计需求](#5-数据库设计需求)
6. [后端功能需求](#6-后端功能需求)
7. [前端应用需求](#7-前端应用需求)
8. [实时通信需求](#8-实时通信需求)
9. [自动派发引擎需求](#9-自动派发引擎需求)
10. [配置与部署需求](#10-配置与部署需求)
11. [非功能性需求](#11-非功能性需求)

---

## 1. 项目背景与目标

### 1.1 背景

开发一套面向餐厅的**扫码点餐全流程管理平台**，覆盖：

- 顾客扫码点餐（移动端，支持同桌多人共享购物车）
- 商家后台管理（PC 端）
- 厨师任务管理（自适应端）
- 服务员送餐管理（移动端）

### 1.2 核心目标

1. **多人实时共享购物车**：同桌多部手机扫码后加入同一购物车会话，任意一人修改购物车，所有人实时同步看到变化。
2. **智能烹饪任务派发**：按时间窗口聚合同一菜品的多张订单，自动匹配菜谱规格（如"宫保鸡丁×2"），分配给负载最低且具备对应技能的厨师。
3. **全流程状态追踪**：从点餐 → 出单 → 烹饪 → 上菜 → 收款，每个环节状态可追踪。
4. **实时推送**：厨师出餐后，服务员端实时收到送餐任务推送。
5. **多区域管理**：餐厅可划分多个用餐区（大厅、包间等），桌台与服务员按区域管理。

---

## 2. 系统角色

| 角色 | 说明 |
|------|------|
| **顾客** | 扫描桌上二维码进入点餐页，添加菜品到购物车，提交订单，查看订单状态 |
| **商家/管理员** | 管理餐桌、菜品、订单、厨师、服务员、区域、系统配置；手动或自动派发烹饪任务 |
| **厨师** | 登录厨师端，查看并接受分配给自己的烹饪任务，标记任务完成 |
| **服务员** | 登录服务员端，通过 WebSocket 实时接收送餐任务，接单、送达或拒绝 |

---

## 3. 技术选型要求

### 3.1 后端

- **语言**：Go 1.21+
- **HTTP 框架**：Gin（`github.com/gin-gonic/gin`）
- **ORM**：GORM + MySQL 驱动（`gorm.io/gorm`，`gorm.io/driver/mysql`）
- **WebSocket**：gorilla/websocket
- **配置**：YAML 文件 + 环境变量双通道，环境变量优先级高于 YAML
- **数据库**：MySQL 8.0

### 3.2 前端（共4个独立应用）

| 应用 | 框架 | UI 库 | 目标端 |
|------|------|-------|--------|
| customer-app（顾客端）| Vue 3 + Vite | Vant 4 | 移动端 H5 |
| merchant-app（商家端）| Vue 3 + Vite | Vuetify 3 | PC 端 |
| chef-app（厨师端）| Vue 3 + Vite | Vuetify 3 | 自适应（平板/PC）|
| waiter-app（服务员端）| Vue 3 + Vite | Vuetify 3 | 移动端自适应 |

每个前端使用 `.env.example` 提供环境变量模板，关键变量：`VITE_API_BASE_URL`。

### 3.3 部署

提供 `docker-compose.yml`，所有服务（MySQL + 后端 + 4个前端）均容器化，执行 `docker-compose up -d` 即可一键启动。

---

## 4. 整体架构要求

```
customer-app (3001) ─┐
merchant-app (3002) ─┤
chef-app     (3003) ─┼──► 后端 API (8080) ──► MySQL (3306)
waiter-app   (3004) ─┘
                      └──► WebSocket /ws/*
```

后端统一提供 REST API 和 WebSocket，前端通过 `VITE_API_BASE_URL` 配置连接地址。

---

## 5. 数据库设计需求

### 5.1 表清单（共14张表）

以下为所有表的字段定义，使用 GORM AutoMigrate 自动建表。

---

#### `dining_area`（用餐区域）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| area_id | INT | PK AUTO | 区域 ID |
| name | VARCHAR(100) | NOT NULL UNIQUE | 区域名称（如：大厅、包间A）|
| description | VARCHAR(255) | | 描述 |
| sort_order | INT | DEFAULT 0 | 排序权重 |
| created_at | DATETIME | autoCreateTime | |

---

#### `table`（餐桌）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| table_id | INT | PK AUTO | |
| table_no | VARCHAR(50) | NOT NULL UNIQUE | 桌号（如 A01）|
| status | ENUM | DEFAULT 'idle' | `idle`/`ordering`/`waiting`/`dining`/`checkout` |
| qr_code | VARCHAR(255) | | 二维码内容/URL |
| capacity | INT | DEFAULT 4 | 座位数 |
| area_id | INT | FK dining_area NULL | 所属区域 |
| created_at | DATETIME | autoCreateTime | |

**桌态流转**：`idle → ordering → waiting → dining → checkout → idle`

---

#### `category`（菜品分类）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| category_id | INT | PK AUTO | |
| name | VARCHAR(100) | NOT NULL | 分类名 |
| sort_order | INT | DEFAULT 0 | 排序 |
| created_at | DATETIME | autoCreateTime | |

---

#### `dish`（菜品）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| dish_id | INT | PK AUTO | |
| category_id | INT | FK category | 所属分类 |
| name | VARCHAR(100) | NOT NULL | 菜品名 |
| price | DECIMAL(10,2) | NOT NULL | 售价 |
| image | VARCHAR(255) | | 图片 URL（JSON 字段名 `image_url`）|
| description | TEXT | | 菜品描述 |
| special_flag | BOOLEAN | DEFAULT false | 是否特色菜 |
| allow_combine | BOOLEAN | DEFAULT true | 是否允许合并烹饪（影响派发）|
| is_available | BOOLEAN | DEFAULT true | 是否上架 |
| sort_order | INT | DEFAULT 0 | 排序 |
| created_at | DATETIME | autoCreateTime | |

---

#### `recipe`（菜谱规格）

一道菜品可有多个菜谱规格，代表不同的烹饪批量单位。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| recipe_id | INT | PK AUTO | |
| dish_id | INT | FK dish NOT NULL | 所属菜品 |
| name | VARCHAR(100) | NOT NULL | 规格名称（如"宫保鸡丁×2"）|
| portion | INT | NOT NULL | 该规格对应份数（如 2）|
| max_portion | INT | DEFAULT 10 | 单批最大份数上限 |
| is_enabled | BOOLEAN | DEFAULT true | 是否启用 |
| created_at | DATETIME | autoCreateTime | |

---

#### `chef`（厨师）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| chef_id | INT | PK AUTO | |
| name | VARCHAR(100) | NOT NULL | 姓名 |
| max_load | INT | DEFAULT 10 | 最大负载份数 |
| current_load | INT | DEFAULT 0 | 当前负载份数 |
| is_active | BOOLEAN | DEFAULT true | 是否在职 |
| created_at | DATETIME | autoCreateTime | |

厨师与菜谱规格是多对多关系（通过 `chef_recipe` 中间表），代表该厨师掌握哪些规格的烹饪技能。

---

#### `chef_recipe`（厨师-菜谱技能关联）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | INT | PK AUTO | |
| chef_id | INT | NOT NULL | |
| recipe_id | INT | NOT NULL | |

联合唯一索引：`(chef_id, recipe_id)`

---

#### `cart_session`（购物车会话）

每张桌台对应一个 active 会话，用于同桌多人共享购物车。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| session_id | INT | PK AUTO | |
| table_id | INT | FK table NOT NULL | |
| status | ENUM | DEFAULT 'active' | `active`/`submitted`/`closed` |
| created_at | DATETIME | autoCreateTime | |

---

#### `session_item`（购物车项）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| item_id | INT | PK AUTO | |
| session_id | INT | FK cart_session NOT NULL | |
| dish_id | INT | FK dish NOT NULL | |
| quantity | INT | DEFAULT 1 | 数量 |
| note | VARCHAR(255) | | 备注（如"不要葱"）|
| added_by | VARCHAR(50) | | 添加者标识（手机标识或用户名）|
| created_at | DATETIME | autoCreateTime | |

---

#### `order`（订单）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| order_id | INT | PK AUTO | |
| table_id | INT | FK table NOT NULL | |
| session_id | INT | FK cart_session NULL | 来源会话 |
| total_amount | DECIMAL(10,2) | | 订单金额 |
| status | ENUM | DEFAULT 'pending' | `pending`/`cooking`/`dining`/`completed`/`cancelled` |
| paid_at | DATETIME | NULL | 收款时间 |
| payment_method | VARCHAR(50) | | 支付方式（如 wechat/cash）|
| is_vip | BOOLEAN | DEFAULT false | VIP 订单（影响派发优先级）|
| created_at | DATETIME | autoCreateTime | |

---

#### `order_item`（订单项）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| item_id | INT | PK AUTO | |
| order_id | INT | FK order NOT NULL | |
| dish_id | INT | FK dish NOT NULL | |
| quantity | INT | DEFAULT 1 | |
| note | VARCHAR(255) | | 备注 |
| status | ENUM | DEFAULT 'pending' | `pending`/`dispatched`/`cooking`/`done`/`served` |
| created_at | DATETIME | autoCreateTime | |

**订单项状态流转**：`pending → dispatched → cooking → done → served`

---

#### `cooking_task`（烹饪任务）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| task_id | INT | PK AUTO | |
| chef_id | INT | FK chef NOT NULL | 分配厨师 |
| recipe_id | INT | FK recipe NOT NULL | 使用菜谱规格 |
| dish_id | INT | FK dish NOT NULL | 菜品 |
| total_portion | INT | NOT NULL | 合计份数 |
| merged_from | JSON | | 源 order_item_id 数组 |
| table_ids | JSON | | 涉及桌号 ID 数组 |
| status | ENUM | DEFAULT 'pending' | `pending`/`cooking`/`done`/`cancelled` |
| priority | INT | DEFAULT 0 | 优先级（VIP=10，普通=0）|
| created_at | DATETIME | autoCreateTime | |
| completed_at | DATETIME | NULL | 完成时间 |

---

#### `system_config`（系统配置）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| config_key | VARCHAR(100) | PK | 配置键 |
| config_value | VARCHAR(500) | NOT NULL | 配置值 |
| description | VARCHAR(255) | | 说明 |

---

#### `waiter`（服务员）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| waiter_id | INT | PK AUTO | |
| name | VARCHAR(100) | NOT NULL | 姓名 |
| phone | VARCHAR(20) | | 联系电话 |
| area_id | INT | FK dining_area NULL | 负责区域 |
| status | ENUM | DEFAULT 'offline' | `online`/`offline` |
| created_at | DATETIME | autoCreateTime | |

服务员连接 WebSocket 后自动变为 `online`，断开后自动变为 `offline`。

---

#### `delivery_task`（送餐任务）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| task_id | INT | PK AUTO | |
| cooking_task_id | INT | FK cooking_task NOT NULL | 关联烹饪任务 |
| waiter_id | INT | FK waiter NULL | 接单服务员 |
| table_ids | JSON | | 目标桌号 ID 数组 |
| status | ENUM | DEFAULT 'pending' | `pending`/`delivering`/`done`/`rejected` |
| pickup_time | DATETIME | NULL | 取餐时间 |
| delivered_time | DATETIME | NULL | 送达时间 |
| reject_reason | VARCHAR(255) | | 拒绝原因 |
| created_at | DATETIME | autoCreateTime | |

---

## 6. 后端功能需求

### 6.1 项目结构

```
backend/
├── main.go
├── config/config.go          # 配置加载
├── database/db.go            # 数据库连接
├── database/migrations.go    # AutoMigrate
├── models/                   # GORM 数据模型
├── handlers/                 # HTTP 处理器（薄层，调用 services）
├── services/                 # 业务逻辑层
├── middleware/cors.go        # CORS 中间件（开发环境允许所有来源）
├── router/router.go          # 路由注册
├── Dockerfile
└── go.mod
```

### 6.2 配置加载规则

配置来源优先级（高→低）：**环境变量 > config.yaml > 代码默认值**

**config.yaml 结构：**

```yaml
server:
  port: "8080"
database:
  host: "localhost"
  port: "3306"
  user: "root"
  password: ""
  name: "restaurant"
dispatch:
  interval_sec: 30     # 自动派发检查间隔（秒）
  window_min: 5        # 订单聚合时间窗口（分钟）
```

**对应环境变量：**`SERVER_PORT`、`DB_HOST`、`DB_PORT`、`DB_USER`、`DB_PASSWORD`、`DB_NAME`、`DISPATCH_INTERVAL_SEC`、`DISPATCH_WINDOW_MIN`、`CONFIG_FILE`（配置文件路径）

### 6.3 启动流程

`main.go` 依次执行：

1. 加载配置
2. 初始化数据库（GORM）
3. 执行 AutoMigrate（含 ALTER TABLE 兼容旧数据库的前置语句）
4. 启动自动派发后台 goroutine（`go RunAutoDispatch(intervalSec, windowMin)`）
5. 启动 Gin HTTP 服务器

### 6.4 所有 API 路由定义

#### 顾客端 `/api/customer`

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/table/:tableId` | 获取餐桌信息 |
| POST | `/cart/session` | 创建或获取当前会话（body: `{table_id}`；若该桌已有 active 会话则返回现有）|
| GET | `/cart/session/:sessionId` | 获取会话（含 items 及 Dish 信息）|
| POST | `/cart/session/:sessionId/item` | 添加菜品（body: `{dish_id, quantity, note, added_by}`；广播 WS）|
| PUT | `/cart/session/:sessionId/item/:itemId` | 修改数量（body: `{quantity}`；广播 WS）|
| DELETE | `/cart/session/:sessionId/item/:itemId` | 删除购物车项（广播 WS）|
| POST | `/cart/session/:sessionId/submit` | 提交订单（body 可选 `{is_vip}`；生成 Order + OrderItem；桌态→waiting）|
| GET | `/order/:orderId` | 查询订单状态（含 items 及 Dish）|
| GET | `/dishes` | 获取全部上架菜品 |
| GET | `/categories` | 获取全部分类 |
| GET | `/dishes/category/:categoryId` | 按分类获取菜品 |

#### 商家端 `/api/merchant`

**餐桌管理：**

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/tables` | 全部餐桌（Preload Area）|
| POST | `/tables` | 新建餐桌 |
| PUT | `/tables/:tableId` | 修改餐桌 |
| DELETE | `/tables/:tableId` | 删除餐桌 |
| GET | `/tables/:tableId/order` | 查看该桌当前订单 |
| PUT | `/tables/:tableId/status` | 手动更新桌态（body: `{status}`）|

**订单管理：**

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/orders` | 订单列表（支持 `?status=` 过滤）|
| GET | `/orders/:orderId` | 订单详情（含 items、Dish、Table）|
| PUT | `/orders/:orderId/status` | 更新订单状态（body: `{status}`）|
| POST | `/orders/:orderId/payment` | 记录收款（body: `{payment_method, discount}`；discount 为折扣比例，如 0.9；更新 paid_at 和 total_amount；桌态→idle）|
| POST | `/orders/:orderId/dispatch` | 手动派发整单（对所有 pending 的 order_item 派发）|
| POST | `/orders/:orderId/items/:itemId/dispatch` | 手动派发单品（body 可选 `{chef_id}`）|

**厨师管理：**

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/chefs` | 厨师列表（含 Recipes）|
| POST | `/chefs` | 新建厨师 |
| PUT | `/chefs/:chefId` | 修改厨师（含技能 recipes 列表同步到 chef_recipe）|
| GET | `/chefs/:chefId/tasks` | 查看某厨师的任务列表 |

**菜谱管理：**

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/recipes` | 全部菜谱（含 Dish）|
| POST | `/recipes` | 新建菜谱规格 |
| PUT | `/recipes/:recipeId` | 修改菜谱 |
| DELETE | `/recipes/:recipeId` | 删除菜谱 |

**菜品管理：**

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/dishes` | 菜品列表（含 Category）|
| POST | `/dishes` | 新建菜品 |
| PUT | `/dishes/:dishId` | 修改菜品 |
| DELETE | `/dishes/:dishId` | 删除菜品 |

**分类管理：**

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/categories` | 分类列表 |
| POST | `/categories` | 新建分类 |
| PUT | `/categories/:categoryId` | 修改分类 |
| DELETE | `/categories/:categoryId` | 删除分类 |

**烹饪任务管理：**

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/tasks` | 任务列表（含 Chef、Recipe、Dish）|
| PUT | `/tasks/:taskId/reassign` | 重派厨师（body: `{chef_id}`；释放旧厨师负载，增加新厨师负载）|

**服务员管理：**

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/waiters` | 服务员列表（含 Area）|
| POST | `/waiters` | 新建服务员 |
| PUT | `/waiters/:waiterId` | 修改服务员 |
| DELETE | `/waiters/:waiterId` | 删除服务员 |

**区域管理：**

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/areas` | 区域列表 |
| POST | `/areas` | 新建区域 |
| PUT | `/areas/:areaId` | 修改区域 |
| DELETE | `/areas/:areaId` | 删除区域 |

**送餐任务管理：**

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/delivery-tasks` | 送餐任务列表（含 CookingTask、Waiter）|
| PUT | `/delivery-tasks/:taskId/return` | 退回送餐任务（重置为 pending，清除 waiter_id）|

**系统配置：**

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | `/config` | 获取全部系统配置（返回 `map[string]string`）|
| PUT | `/config` | 更新系统配置（body: `map[string]string`，批量 upsert）|

#### 厨师端 `/api/chef`

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | `/auth/login` | 登录（body: `{chef_id}`；简化认证，返回厨师信息）|
| GET | `/chefs/:chefId` | 获取厨师信息 |
| GET | `/tasks` | 获取任务列表（`?chef_id=` 过滤，含 Recipe、Dish；默认返回 pending+cooking 状态）|
| GET | `/tasks/:taskId` | 任务详情 |
| PUT | `/tasks/:taskId/complete` | 标记任务完成（详见业务规则 §6.5）|

#### 服务员端 `/api/waiter`

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | `/auth/login` | 登录（body: `{waiter_id}`；简化认证）|
| GET | `/waiters/:waiterId` | 获取服务员信息 |
| GET | `/tasks` | 获取送餐任务（`?waiter_id=` 过滤；默认返回 pending+delivering 状态）|
| GET | `/tasks/:taskId` | 任务详情（含 Tables）|
| PUT | `/tasks/:taskId/pickup` | 接单取餐（status→delivering，记录 pickup_time）|
| PUT | `/tasks/:taskId/deliver` | 确认送达（status→done，记录 delivered_time；详见业务规则 §6.5）|
| PUT | `/tasks/:taskId/reject` | 拒绝任务（body: `{reason}`；status→rejected，记录 reject_reason；重新生成 pending 送餐任务并广播）|

#### WebSocket

| 路径 | 说明 |
|------|------|
| `GET /ws/cart/:sessionId` | 购物车同步（任意购物车修改后广播完整 CartSession 快照）|
| `GET /ws/waiter/:waiterId` | 服务员推送（连接时 status→online；断开时 status→offline）|

### 6.5 关键业务规则

#### 任务完成级联（`PUT /chef/tasks/:taskId/complete`）

1. `cooking_task.status = 'done'`，`completed_at = NOW()`
2. `merged_from` 中所有 `order_item.status = 'done'`
3. `chef.current_load -= task.total_portion`（不低于 0）
4. 自动创建 `delivery_task`（`status='pending'`，复制 `table_ids`）
5. 通过 WaiterHub 广播新送餐任务给对应区域在线服务员（消息格式见 §8）
6. 检查关联订单：若该订单所有 order_item 均为 `done` 或 `served`，则 `order.status = 'dining'`

#### 送达确认级联（`PUT /waiter/tasks/:taskId/deliver`）

1. `delivery_task.status = 'done'`，`delivered_time = NOW()`
2. 关联 `order_item.status = 'served'`
3. 检查关联订单：若所有 order_item 均为 `served`，则 `order.status = 'dining'`

#### 收款级联（`POST /merchant/orders/:orderId/payment`）

1. 按 `discount` 折扣比例重算 `total_amount`
2. `order.payment_method`、`order.paid_at = NOW()`
3. `order.status = 'completed'`
4. `table.status = 'idle'`

---

## 7. 前端应用需求

### 7.1 顾客端（customer-app）端口 3001

**路由：**

| 路径 | 组件 | 说明 |
|------|------|------|
| `/order/:tableId` | OrderPage | 点餐主页（扫码后进入，`tableId` 从 URL 参数获取）|
| `/status/:orderId` | OrderStatus | 订单状态页 |

**OrderPage 功能：**

- 进入页面时以 `tableId` 调用 `POST /api/customer/cart/session` 获取/创建会话
- 连接 WebSocket `ws://{HOST}/ws/cart/{sessionId}`，接收到消息时刷新购物车展示
- 展示菜品分类标签页（调用 `/api/customer/categories` 和 `/api/customer/dishes`）
- 支持按分类筛选菜品
- 菜品卡片显示：名称、价格、图片、描述；加号/减号按钮操作数量
- 添加菜品时可填写备注（note）
- 底部购物车栏显示当前总价和菜品数量，点击展开购物车列表
- 购物车列表可删除/修改菜品数量
- 点击"提交订单"按钮：调用 `POST .../submit`，成功后跳转到 `/status/:orderId`

**OrderStatus 功能：**

- 轮询（或 WS）展示订单状态和每个 order_item 的状态
- 展示订单总金额、桌号
- 状态以可视化标签展示

---

### 7.2 商家端（merchant-app）端口 3002

全局导航侧边栏，包含以下模块：

#### 餐桌管理（TablesView）

- 网格布局展示所有餐桌，卡片颜色按桌态区分：
  - `idle`=绿色，`ordering`=蓝色，`waiting`=橙色，`dining`=红色，`checkout`=紫色
- 点击餐桌卡片：弹出该桌当前订单详情（含 order_item 状态）
- 支持新增、编辑、删除餐桌（包含桌号、座位数、所属区域字段）
- 支持手动更新桌态

#### 订单管理（OrdersView）

- 列表展示所有订单，支持按状态筛选
- 订单行显示：桌号、金额、状态、创建时间、VIP 标记
- 点击订单：展开详情（含 order_item 列表及每项状态）
- 支持收款操作（选择支付方式、输入折扣）
- 支持整单手动派发和单品派发（可指定厨师）

#### 烹饪任务（TasksView）

- 任务列表：显示厨师名、菜品名、规格、合计份数、涉及桌号、状态
- 支持将任务重新派发给其他厨师

#### 送餐任务（DeliveryTasksView）

- 送餐任务列表：显示关联烹饪任务、服务员、桌号、状态、时间
- 支持退回任务（重置为 pending）

#### 菜品管理（DishesView）

- 菜品 CRUD，字段：名称、分类、价格、图片、描述、是否特色、是否上架、是否允许合并烹饪

#### 分类管理（CategoriesView）

- 分类 CRUD，字段：名称、排序

#### 菜谱管理（RecipesView）

- 菜谱规格 CRUD，字段：关联菜品（下拉选择）、规格名称、份数、最大份数、是否启用

#### 厨师管理（ChefsView）

- 厨师 CRUD，字段：姓名、最大负载
- 支持为厨师分配菜谱技能（多选 recipe，写入 chef_recipe 表）
- 显示当前负载

#### 服务员管理（WaitersView）

- 服务员 CRUD，字段：姓名、电话、负责区域（下拉选 dining_area）
- 显示在线状态

#### 区域管理（AreasView）

- 区域 CRUD，字段：名称、描述、排序

#### 系统配置（ConfigView）

- 表单形式展示和编辑 system_config 中的配置项

---

### 7.3 厨师端（chef-app）端口 3003

**登录页（LoginView）：**

- 下拉列表选择厨师姓名（调用 `GET /api/merchant/chefs`）
- 点击登录：调用 `POST /api/chef/auth/login`，存储 `chef_id` 到 localStorage
- 登录后跳转任务列表

**任务列表（TasksView）：**

- 轮询或手动刷新，展示当前厨师的任务列表
- 每张任务卡片显示：
  - 菜品名称
  - 菜谱规格名称（如"宫保鸡丁×2"）
  - 合计份数
  - 涉及桌号（多桌合并时显示所有桌号）
  - 任务状态
  - 创建时间
- 对 `pending` 状态任务显示"开始烹饪"按钮（可选，更新为 `cooking`）
- 对 `pending`/`cooking` 状态任务显示"完成"按钮 → 调用 `PUT /api/chef/tasks/:taskId/complete`
- 登出按钮（清除 localStorage）

---

### 7.4 服务员端（waiter-app）端口 3004

**登录页（LoginView）：**

- 下拉列表选择服务员姓名（调用 `GET /api/merchant/waiters`）
- 点击登录：调用 `POST /api/waiter/auth/login`，存储 `waiter_id` 到 localStorage
- 登录后跳转任务列表，**同时建立 WebSocket 连接** `ws://{HOST}/ws/waiter/{waiterId}`

**任务列表（TasksView）：**

- 展示送餐任务列表（pending + delivering 状态）
- WebSocket 收到新任务推送时，自动在列表顶部显示新任务（高亮提示）
- 每张任务卡片显示：目标桌号（可能多桌）、菜品名称、状态、时间
- 操作按钮：
  - `pending` 状态：**"接单取餐"** → 调用 `PUT .../pickup`
  - `delivering` 状态：**"确认送达"** → 调用 `PUT .../deliver`；**"拒绝"**（填写原因）→ 调用 `PUT .../reject`
- 登出按钮（清除 localStorage，断开 WebSocket）

---

## 8. 实时通信需求

### 8.1 购物车 WebSocket Hub

- 按 `sessionId` 分组管理连接（`map[int][]*WSClient`）
- 购物车任意变化（增/删/改）后，向该 sessionId 所有客户端广播**完整 CartSession JSON 快照**（含 items 和每个 item 的 Dish 信息）
- 客户端断开连接时自动从 Hub 移除

**消息格式（服务端 → 客户端）：**
```json
{
  "session_id": 5,
  "table_id": 2,
  "status": "active",
  "created_at": "...",
  "items": [
    {
      "item_id": 12,
      "dish_id": 3,
      "quantity": 2,
      "note": "不要辣",
      "added_by": "用户A",
      "dish": { "dish_id": 3, "name": "宫保鸡丁", "price": 28.00 }
    }
  ]
}
```

### 8.2 服务员 WebSocket Hub

- 按 `waiterId` 管理连接（`map[int]*WaiterWSClient`，每个服务员只有一个连接）
- 厨师完成任务时自动创建 delivery_task，并向**符合区域条件**的在线服务员推送
- **区域过滤规则**：
  - 服务员 `area_id = 0`（未设置）→ 接收所有区域任务
  - 服务员 `area_id = N` → 只接收 `area_id = N` 的任务

**消息格式（服务端 → 客户端）：**
```json
{
  "type": "new_delivery_task",
  "task": {
    "task_id": 8,
    "cooking_task_id": 12,
    "table_ids": [3, 5],
    "status": "pending",
    "created_at": "..."
  }
}
```

---

## 9. 自动派发引擎需求

### 9.1 后台定时触发

- 启动一个后台 goroutine（`RunAutoDispatch`）
- 每隔 `DISPATCH_INTERVAL_SEC` 秒（默认 30）执行一次 `AutoDispatch`

### 9.2 AutoDispatch 逻辑

1. 查询所有满足以下条件的 `order_item`：
   - `status = 'pending'`
   - `created_at >= NOW() - DISPATCH_WINDOW_MIN 分钟`
2. 按 `dish_id` 分组，汇总：
   - `total_qty`（所有 item 的 quantity 之和）
   - `item_ids`（源 item_id 列表，存入 `merged_from`）
   - `table_ids`（涉及的桌号 ID 集合）
   - `has_vip`（是否有 VIP 订单）
3. 对每个 dish_id 分组，调用 `createTasksForDish`

### 9.3 菜谱选择算法（`findBestRecipe`）

按优先级选择最优菜谱规格（`total_qty` 为目标份数）：

1. **精确匹配**：`portion == total_qty`
2. **最小覆盖**：最小的 `portion >= total_qty`（避免浪费，用最合适的规格覆盖）
3. **组合拆分**（仅 `allow_combine = true` 时）：最大的 `portion <= total_qty`
4. **兜底**：`portion` 最大的规格

### 9.4 厨师分配算法（`assignChef`）

```sql
SELECT * FROM chef
JOIN chef_recipe cr ON cr.chef_id = chef.chef_id AND cr.recipe_id = ?
WHERE chef.is_active = 1
  AND chef.current_load < chef.max_load
ORDER BY chef.current_load ASC
LIMIT 1
```

分配后：`chef.current_load += total_portion`

### 9.5 任务创建后的更新

- `order_item.status = 'dispatched'`（所有参与合并的 item）
- `cooking_task` 中记录 `merged_from`（item_id 数组）和 `table_ids`（桌号数组，已排序去重）
- VIP 订单 `priority = 10`，普通订单 `priority = 0`

### 9.6 手动派发

- **整单派发**（`POST /merchant/orders/:orderId/dispatch`）：对该订单内所有 `pending` 状态的 `order_item` 逐个 dish 派发
- **单品派发**（`POST /merchant/orders/:orderId/items/:itemId/dispatch`）：可通过 body 指定 `chef_id`（覆盖自动分配）

---

## 10. 配置与部署需求

### 10.1 Docker Compose

提供 `docker-compose.yml`，包含以下 service：

| Service | 镜像/Build | 端口映射 | 依赖 |
|---------|-----------|---------|------|
| mysql | mysql:8.0 | 3306:3306 | — |
| backend | ./backend/Dockerfile | 8080:8080 | mysql（healthcheck）|
| customer-app | ./customer-app/Dockerfile | 3001:80 | backend |
| merchant-app | ./merchant-app/Dockerfile | 3002:80 | backend |
| chef-app | ./chef-app/Dockerfile | 3003:80 | backend |
| waiter-app | ./waiter-app/Dockerfile | 3004:80 | backend |

MySQL 配置：
- `MYSQL_DATABASE=restaurant`
- `MYSQL_USER=restaurant`
- `MYSQL_PASSWORD=restaurant123`
- `MYSQL_ROOT_PASSWORD=rootpassword`

Backend 通过环境变量接收数据库连接信息和派发参数。

### 10.2 各应用 Dockerfile

**后端 Dockerfile** 要求：
- 多阶段构建：`golang:1.21` builder → `debian:bookworm-slim` 运行镜像
- 编译产出单一二进制文件，暴露 8080 端口

**前端 Dockerfile** 要求：
- 多阶段构建：`node:18-alpine` builder (`npm run build`) → `nginx:alpine` 静态文件服务
- 提供 `nginx.conf`，将所有路径回退到 `index.html`（支持 Vue Router history/hash 模式）

### 10.3 前端环境变量

每个前端应用提供 `.env.example`：
```
VITE_API_BASE_URL=http://localhost:8080
```

---

## 11. 非功能性需求

| 类别 | 要求 |
|------|------|
| **跨域** | 后端 CORS 中间件允许所有来源（开发环境）；生产应限定前端域名 |
| **认证** | 厨师端和服务员端采用简化认证（直接传入 ID，无密码）；生产环境可扩展 JWT |
| **错误处理** | HTTP 错误均以 `{"error": "..."}` JSON 格式返回；404、400、500 统一处理 |
| **数据库迁移** | 使用 GORM AutoMigrate + 前置 ALTER TABLE 语句保证兼容升级（新增列用 `ADD COLUMN IF NOT EXISTS`）|
| **并发安全** | WebSocket Hub 使用 `sync.RWMutex` 保护客户端 map |
| **日志** | 使用 Go 标准库 `log` 输出关键操作日志（派发结果、错误）|
| **数据一致性** | 厨师负载增减操作使用 SQL 表达式（`GREATEST(0, current_load - N)`）防止负值 |

---

## 附录：关键数据流示例

### A. 顾客点餐到派发完成

```
1. 顾客扫码 → GET /api/customer/table/1
2. POST /api/customer/cart/session  {table_id: 1}
   → 返回 session_id=5
3. WS 连接 ws://{HOST}/ws/cart/5
4. POST /api/customer/cart/session/5/item  {dish_id: 3, quantity: 2}
   → WS 广播更新后的 CartSession 给所有同会话客户端
5. POST /api/customer/cart/session/5/submit
   → 生成 order_id=10，order_item(dish_id=3, qty=2, status=pending)
   → table.status = 'waiting'
6. （30秒后）AutoDispatch 执行：
   → 查找 order_item status=pending，时间窗口内
   → 按 dish_id 分组：dish_id=3, total_qty=2
   → findBestRecipe → 匹配 portion=2 的菜谱 recipe_id=7
   → assignChef → 分配负载最低的厨师 chef_id=2
   → 创建 cooking_task，order_item.status=dispatched
7. 厨师端：GET /api/chef/tasks?chef_id=2 → 看到新任务
8. PUT /api/chef/tasks/15/complete
   → order_item.status=done
   → 创建 delivery_task
   → WS 推送给服务员
9. 服务员：PUT /api/waiter/tasks/20/pickup → delivering
10. PUT /api/waiter/tasks/20/deliver → done，order_item.status=served
11. 商家：POST /api/merchant/orders/10/payment {payment_method:'wechat', discount:1.0}
    → order.status=completed，table.status=idle
```

### B. 多人共享购物车

```
手机A: POST /cart/session {table_id:1} → session_id=5, WS连接
手机B: POST /cart/session {table_id:1} → 返回同一个 session_id=5, WS连接
手机A: POST /cart/session/5/item {dish_id:1, qty:1}
  → 服务端: 写入 session_item
  → WS广播: 手机A 和 手机B 同时收到最新 CartSession
手机B看到手机A添加的菜，再添加自己的菜
  → WS广播: 双方同步
```
