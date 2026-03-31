# 餐厅扫码点餐系统 — 完整项目文档

> 版本：2026-03 | 联系：893400722@qq.com

---

## 目录

1. [项目概述](#1-项目概述)
2. [系统架构](#2-系统架构)
3. [技术栈](#3-技术栈)
4. [项目目录结构](#4-项目目录结构)
5. [数据库设计](#5-数据库设计)
6. [后端服务](#6-后端服务)
7. [API 接口文档](#7-api-接口文档)
8. [WebSocket 协议](#8-websocket-协议)
9. [自动派发引擎](#9-自动派发引擎)
10. [前端应用](#10-前端应用)
11. [配置说明](#11-配置说明)
12. [快速启动](#12-快速启动)
13. [本地开发](#13-本地开发)
14. [常见问题](#14-常见问题)

---

## 1. 项目概述

本系统是一套完整的**餐厅扫码点餐平台**，支持多客户端实时协同，覆盖从顾客点餐到厨师出餐、服务员上菜的完整餐饮业务流程。

### 核心亮点

| 特性 | 说明 |
|------|------|
| **多人共享购物车** | 同桌多部手机通过 WebSocket 实时同步，一张桌一个购物车 |
| **智能烹饪派发** | 按时间窗口聚合订单，自动匹配菜谱规格并分配最优厨师 |
| **全流程追踪** | 订单状态：待处理 → 烹饪中 → 用餐中 → 已完成 |
| **送餐任务管理** | 厨师完成后自动生成送餐任务，服务员端实时推送接收 |
| **多区域管理** | 餐厅可划分多个用餐区，服务员、餐桌按区域管理 |
| **Docker 一键部署** | 所有服务容器化，`docker-compose up` 即可运行 |

### 业务流程总览

```
顾客扫码 → 创建/加入购物车会话 → 多人协同添加菜品
    → 提交订单 → 自动/手动派发烹饪任务
    → 厨师接单烹饪 → 完成后生成送餐任务
    → 服务员接单 → 上菜确认
    → 商家收款结账
```

---

## 2. 系统架构

```
┌──────────────────────────────────────────────────────────────┐
│                         客户端层                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──┐ │
│  │  顾客端       │  │   商家端      │  │   厨师端      │  │服│ │
│  │ customer-app │  │ merchant-app │  │   chef-app   │  │务│ │
│  │  Vue3+Vant   │  │Vue3+Vuetify  │  │Vue3+Vuetify  │  │员│ │
│  │  Port:3001   │  │  Port:3002   │  │  Port:3003   │  │端│ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └┬─┘ │
└─────────┼────────────────┼────────────────┼─────────────┼───┘
          │                │                │             │
          ▼                ▼                ▼             ▼
┌──────────────────────────────────────────────────────────────┐
│                    后端 API 层 (Port: 8080)                    │
│                    Go + Gin Framework                         │
│                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌────────────────────┐   │
│  │ REST API    │  │  WebSocket  │  │  自动派发引擎        │   │
│  │ /api/*      │  │ /ws/cart/*  │  │  (后台 goroutine)  │   │
│  │             │  │ /ws/waiter/*│  │                    │   │
│  └─────────────┘  └─────────────┘  └────────────────────┘   │
└──────────────────────────────┬───────────────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────┐
│                    数据库层 (Port: 3306)                       │
│                      MySQL 8.0                                │
└──────────────────────────────────────────────────────────────┘
```

---

## 3. 技术栈

### 后端

| 组件 | 版本 | 用途 |
|------|------|------|
| Go | 1.21 | 运行时 |
| Gin | v1.9.1 | HTTP 框架 / 路由 |
| GORM | v1.30.0 | ORM（MySQL 驱动）|
| gorilla/websocket | v1.5.1 | WebSocket 实时通信 |
| gopkg.in/yaml.v3 | v3.0.1 | YAML 配置文件解析 |
| MySQL | 8.0 | 关系型数据库 |

### 前端（顾客端）

| 组件 | 版本 | 用途 |
|------|------|------|
| Vue 3 | - | 响应式框架 |
| Vant 4 | - | 移动端 UI 组件库 |
| Vue Router | - | 客户端路由 |
| Pinia | - | 状态管理 |
| Vite | - | 构建工具 |

### 前端（商家端 / 厨师端 / 服务员端）

| 组件 | 版本 | 用途 |
|------|------|------|
| Vue 3 | - | 响应式框架 |
| Vuetify 3 | - | Material Design UI |
| Vue Router | - | 客户端路由 |
| Vite | - | 构建工具 |

---

## 4. 项目目录结构

```
chengtb/
├── backend/                    # Go 后端服务
│   ├── main.go                 # 程序入口
│   ├── config/
│   │   └── config.go           # 配置加载（YAML + 环境变量）
│   ├── database/
│   │   ├── db.go               # GORM 数据库连接
│   │   └── migrations.go       # 自动迁移
│   ├── models/                 # 数据模型（GORM struct）
│   │   ├── area.go             # 用餐区域
│   │   ├── cart.go             # 购物车会话 & 会话项
│   │   ├── category.go         # 菜品分类
│   │   ├── chef.go             # 厨师 & 厨师-菜谱关联
│   │   ├── dish.go             # 菜品
│   │   ├── order.go            # 订单 & 订单项
│   │   ├── recipe.go           # 菜谱规格
│   │   ├── table.go            # 餐桌
│   │   ├── task.go             # 烹饪任务 & 系统配置
│   │   └── waiter.go           # 服务员 & 送餐任务
│   ├── handlers/               # HTTP 处理器
│   │   ├── customer.go         # 顾客端接口
│   │   ├── merchant.go         # 商家端接口
│   │   ├── chef.go             # 厨师端接口
│   │   ├── waiter.go           # 服务员端接口
│   │   └── websocket.go        # WebSocket 升级
│   ├── services/               # 业务逻辑层
│   │   ├── cart.go             # 购物车操作
│   │   ├── dispatch.go         # 烹饪任务派发引擎
│   │   ├── order.go            # 订单操作
│   │   └── websocket.go        # WS Hub（购物车 & 服务员）
│   ├── middleware/
│   │   └── cors.go             # CORS 中间件
│   ├── router/
│   │   └── router.go           # 路由注册
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
│
├── customer-app/               # 顾客点餐应用（移动端 H5）
│   ├── src/
│   │   ├── views/
│   │   │   ├── OrderPage.vue   # 点餐主页面
│   │   │   └── OrderStatus.vue # 订单状态页
│   │   ├── api/                # Axios API 封装
│   │   ├── store/              # Pinia 状态
│   │   ├── router/             # Vue Router
│   │   └── components/         # 公共组件
│   ├── Dockerfile
│   └── package.json
│
├── merchant-app/               # 商家管理后台（PC 端）
│   ├── src/
│   │   ├── views/
│   │   │   ├── TablesView.vue      # 餐桌管理
│   │   │   ├── OrdersView.vue      # 订单管理
│   │   │   ├── TasksView.vue       # 烹饪任务
│   │   │   ├── DeliveryTasksView.vue # 送餐任务
│   │   │   ├── DishesView.vue      # 菜品管理
│   │   │   ├── CategoriesView.vue  # 分类管理
│   │   │   ├── RecipesView.vue     # 菜谱管理
│   │   │   ├── ChefsView.vue       # 厨师管理
│   │   │   ├── WaitersView.vue     # 服务员管理
│   │   │   ├── AreasView.vue       # 区域管理
│   │   │   └── ConfigView.vue      # 系统配置
│   │   ├── api/
│   │   └── router/
│   ├── Dockerfile
│   └── package.json
│
├── chef-app/                   # 厨师任务端（自适应）
│   ├── src/
│   │   ├── views/
│   │   │   ├── LoginView.vue   # 厨师登录
│   │   │   └── TasksView.vue   # 任务列表
│   │   ├── api/
│   │   └── router/
│   ├── Dockerfile
│   └── package.json
│
├── waiter-app/                 # 服务员送餐端
│   ├── src/
│   │   ├── views/
│   │   │   ├── LoginView.vue   # 服务员登录
│   │   │   └── TasksView.vue   # 送餐任务列表
│   │   ├── api/
│   │   └── router/
│   ├── Dockerfile
│   └── package.json
│
├── docker-compose.yml          # 一键部署配置
└── README.md
```

---

## 5. 数据库设计

### ER 图（文字描述）

```
dining_area ──── table (area_id)
dining_area ──── waiter (area_id)
table ──────────── cart_session (table_id)
cart_session ──── session_item (session_id)
session_item ──── dish (dish_id)
table ──────────── order (table_id)
cart_session ──── order (session_id)
order ───────────── order_item (order_id)
order_item ───── dish (dish_id)
category ────────── dish (category_id)
dish ────────────── recipe (dish_id)
chef ────────────── chef_recipe ──── recipe
cooking_task ──── chef, recipe, dish
delivery_task ──── cooking_task, waiter
```

### 数据表详情

#### `dining_area`（用餐区域）

| 列名 | 类型 | 说明 |
|------|------|------|
| area_id | INT PK AUTO | 区域 ID |
| name | VARCHAR(100) UNIQUE | 区域名称（如：大厅、包间A）|
| description | VARCHAR(255) | 区域描述 |
| sort_order | INT DEFAULT 0 | 排序 |
| created_at | DATETIME | 创建时间 |

#### `table`（餐桌）

| 列名 | 类型 | 说明 |
|------|------|------|
| table_id | INT PK AUTO | 桌号 ID |
| table_no | VARCHAR(50) UNIQUE | 桌号（如：A01）|
| status | ENUM | `idle`/`ordering`/`waiting`/`dining`/`checkout` |
| qr_code | VARCHAR(255) | 二维码内容 |
| capacity | INT DEFAULT 4 | 最大容纳人数 |
| area_id | INT FK | 所属区域 |
| created_at | DATETIME | 创建时间 |

桌态流转：
```
idle → ordering → waiting → dining → checkout → idle
```

#### `category`（菜品分类）

| 列名 | 类型 | 说明 |
|------|------|------|
| category_id | INT PK AUTO | 分类 ID |
| name | VARCHAR(100) | 分类名 |
| sort_order | INT DEFAULT 0 | 排序 |
| created_at | DATETIME | 创建时间 |

#### `dish`（菜品）

| 列名 | 类型 | 说明 |
|------|------|------|
| dish_id | INT PK AUTO | 菜品 ID |
| category_id | INT FK | 所属分类 |
| name | VARCHAR(100) | 菜品名称 |
| price | DECIMAL(10,2) | 售价 |
| image | VARCHAR(255) | 图片 URL |
| description | TEXT | 菜品描述 |
| special_flag | BOOLEAN | 是否特色菜 |
| allow_combine | BOOLEAN | 是否允许合并烹饪 |
| is_available | BOOLEAN | 是否上架 |
| sort_order | INT DEFAULT 0 | 排序 |
| created_at | DATETIME | 创建时间 |

#### `recipe`（菜谱规格）

| 列名 | 类型 | 说明 |
|------|------|------|
| recipe_id | INT PK AUTO | 菜谱 ID |
| dish_id | INT FK | 关联菜品 |
| name | VARCHAR(100) | 规格名称（如：宫保鸡丁×2）|
| portion | INT | 该规格对应的份数 |
| max_portion | INT DEFAULT 10 | 单批最大份数上限 |
| is_enabled | BOOLEAN | 是否启用 |
| created_at | DATETIME | 创建时间 |

> 一道菜品可有多个菜谱规格（如 x1、x2、x4），派发时自动选最优规格。

#### `chef`（厨师）

| 列名 | 类型 | 说明 |
|------|------|------|
| chef_id | INT PK AUTO | 厨师 ID |
| name | VARCHAR(100) | 姓名 |
| max_load | INT DEFAULT 10 | 最大负载（份）|
| current_load | INT DEFAULT 0 | 当前负载（份）|
| is_active | BOOLEAN | 是否在职 |
| created_at | DATETIME | 创建时间 |

#### `chef_recipe`（厨师技能表）

| 列名 | 类型 | 说明 |
|------|------|------|
| id | INT PK AUTO | |
| chef_id | INT FK | 厨师 |
| recipe_id | INT FK | 可烹饪的菜谱规格 |

> 联合唯一索引 `(chef_id, recipe_id)`

#### `cart_session`（购物车会话）

| 列名 | 类型 | 说明 |
|------|------|------|
| session_id | INT PK AUTO | 会话 ID |
| table_id | INT FK | 餐桌 |
| status | ENUM | `active`/`submitted`/`closed` |
| created_at | DATETIME | 创建时间 |

#### `session_item`（购物车项）

| 列名 | 类型 | 说明 |
|------|------|------|
| item_id | INT PK AUTO | |
| session_id | INT FK | 所属会话 |
| dish_id | INT FK | 菜品 |
| quantity | INT DEFAULT 1 | 数量 |
| note | VARCHAR(255) | 备注（如：不要葱）|
| added_by | VARCHAR(50) | 添加者标识 |
| created_at | DATETIME | 创建时间 |

#### `order`（订单）

| 列名 | 类型 | 说明 |
|------|------|------|
| order_id | INT PK AUTO | 订单 ID |
| table_id | INT FK | 餐桌 |
| session_id | INT FK | 来源会话 |
| total_amount | DECIMAL(10,2) | 订单金额 |
| status | ENUM | `pending`/`cooking`/`dining`/`completed`/`cancelled` |
| paid_at | DATETIME NULL | 收款时间 |
| payment_method | VARCHAR(50) | 支付方式 |
| is_vip | BOOLEAN | 是否 VIP 订单（影响派发优先级）|
| created_at | DATETIME | 创建时间 |

#### `order_item`（订单项）

| 列名 | 类型 | 说明 |
|------|------|------|
| item_id | INT PK AUTO | |
| order_id | INT FK | 所属订单 |
| dish_id | INT FK | 菜品 |
| quantity | INT DEFAULT 1 | 数量 |
| note | VARCHAR(255) | 备注 |
| status | ENUM | `pending`/`dispatched`/`cooking`/`done`/`served` |
| created_at | DATETIME | 创建时间 |

订单项状态流转：
```
pending → dispatched → cooking → done → served
```

#### `cooking_task`（烹饪任务）

| 列名 | 类型 | 说明 |
|------|------|------|
| task_id | INT PK AUTO | 任务 ID |
| chef_id | INT FK | 分配厨师 |
| recipe_id | INT FK | 使用菜谱规格 |
| dish_id | INT FK | 菜品 |
| total_portion | INT | 合计份数 |
| merged_from | JSON | 源 order_item_id 数组（合并来源）|
| table_ids | JSON | 涉及餐桌 ID 数组 |
| status | ENUM | `pending`/`cooking`/`done`/`cancelled` |
| priority | INT DEFAULT 0 | 优先级（VIP=10，普通=0）|
| created_at | DATETIME | 创建时间 |
| completed_at | DATETIME NULL | 完成时间 |

#### `system_config`（系统配置）

| 列名 | 类型 | 说明 |
|------|------|------|
| config_key | VARCHAR(100) PK | 配置键 |
| config_value | VARCHAR(500) | 配置值 |
| description | VARCHAR(255) | 说明 |

#### `waiter`（服务员）

| 列名 | 类型 | 说明 |
|------|------|------|
| waiter_id | INT PK AUTO | |
| name | VARCHAR(100) | 姓名 |
| phone | VARCHAR(20) | 电话 |
| area_id | INT FK NULL | 负责区域 |
| status | ENUM | `online`/`offline` |
| created_at | DATETIME | 创建时间 |

#### `delivery_task`（送餐任务）

| 列名 | 类型 | 说明 |
|------|------|------|
| task_id | INT PK AUTO | |
| cooking_task_id | INT FK | 关联烹饪任务 |
| waiter_id | INT FK NULL | 接单服务员 |
| table_ids | JSON | 目标餐桌 ID 数组 |
| status | ENUM | `pending`/`delivering`/`done`/`rejected` |
| pickup_time | DATETIME NULL | 取餐时间 |
| delivered_time | DATETIME NULL | 送达时间 |
| reject_reason | VARCHAR(255) | 拒绝原因 |
| created_at | DATETIME | 创建时间 |

---

## 6. 后端服务

### 入口与启动流程

`main.go` 启动时依次执行：

1. 加载配置（`config.Load()`）
2. 初始化数据库连接（`database.Init()`）
3. 执行 GORM AutoMigrate（`database.Migrate()`）
4. 启动自动派发后台 goroutine（`services.RunAutoDispatch()`）
5. 初始化 Gin 路由并监听端口

### 中间件

- **CORS**：允许所有来源（开发环境），生产环境应限定前端域名。

### 服务层（services/）

| 文件 | 主要功能 |
|------|---------|
| `cart.go` | 创建/获取会话、增删购物车项、提交订单 |
| `dispatch.go` | 自动/手动派发、最优菜谱选择、厨师分配、任务重派 |
| `order.go` | 订单状态更新、收款记录 |
| `websocket.go` | 购物车 WS Hub、服务员 WS Hub 的注册/广播 |

---

## 7. API 接口文档

> 所有接口 Base URL：`http://localhost:8080`
> Content-Type：`application/json`

---

### 7.1 顾客端 `/api/customer`

#### 获取餐桌信息
```
GET /api/customer/table/:tableId
```
响应：`Table` 对象

---

#### 创建或获取购物车会话
```
POST /api/customer/cart/session
Body: { "table_id": 1 }
```
响应：`CartSession` 对象（含 `items`）

> 若该桌已有 `active` 会话则返回现有会话，否则新建。

---

#### 获取购物车会话
```
GET /api/customer/cart/session/:sessionId
```
响应：`CartSession` 对象（含 `items` 及 `Dish` 信息）

---

#### 添加菜品到购物车
```
POST /api/customer/cart/session/:sessionId/item
Body: {
  "dish_id": 5,
  "quantity": 2,
  "note": "不要葱",
  "added_by": "用户A"
}
```
响应：更新后的 `CartSession`，并通过 WebSocket 广播给同会话所有客户端。

---

#### 修改购物车项数量
```
PUT /api/customer/cart/session/:sessionId/item/:itemId
Body: { "quantity": 3 }
```
响应：更新后的 `CartSession`，并广播。

---

#### 删除购物车项
```
DELETE /api/customer/cart/session/:sessionId/item/:itemId
```
响应：更新后的 `CartSession`，并广播。

---

#### 提交订单
```
POST /api/customer/cart/session/:sessionId/submit
Body: { "is_vip": false }   （可选）
```
响应：新建的 `Order` 对象。

> 将 `CartSession` 中的 `session_item` 转为 `order_item`，桌态更新为 `waiting`。

---

#### 查询订单状态
```
GET /api/customer/order/:orderId
```
响应：`Order` 对象（含 `items`、`Dish` 信息）

---

#### 获取菜单（全部菜品）
```
GET /api/customer/dishes
```
响应：`[]Dish`

---

#### 获取菜品分类
```
GET /api/customer/categories
```
响应：`[]Category`

---

#### 按分类获取菜品
```
GET /api/customer/dishes/category/:categoryId
```
响应：`[]Dish`

---

### 7.2 商家端 `/api/merchant`

#### 餐桌管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/merchant/tables` | 获取全部餐桌（含区域） |
| POST | `/merchant/tables` | 新建餐桌 |
| PUT | `/merchant/tables/:tableId` | 修改餐桌信息 |
| DELETE | `/merchant/tables/:tableId` | 删除餐桌 |
| GET | `/merchant/tables/:tableId/order` | 获取餐桌当前订单 |
| PUT | `/merchant/tables/:tableId/status` | 手动更新桌态 |

---

#### 订单管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/merchant/orders` | 订单列表（支持 status 过滤）|
| GET | `/merchant/orders/:orderId` | 订单详情 |
| PUT | `/merchant/orders/:orderId/status` | 更新订单状态 |
| POST | `/merchant/orders/:orderId/payment` | 记录收款 |
| POST | `/merchant/orders/:orderId/dispatch` | 手动派发整单 |
| POST | `/merchant/orders/:orderId/items/:itemId/dispatch` | 手动派发单个菜品 |

**收款请求体：**
```json
{
  "payment_method": "wechat",
  "discount": 0.9
}
```

**手动派发单品请求体（可选指定厨师）：**
```json
{
  "chef_id": 3
}
```

---

#### 厨师管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/merchant/chefs` | 厨师列表 |
| POST | `/merchant/chefs` | 新建厨师 |
| PUT | `/merchant/chefs/:chefId` | 修改厨师信息 |
| GET | `/merchant/chefs/:chefId/tasks` | 查看某厨师的任务 |

---

#### 菜谱管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/merchant/recipes` | 菜谱列表 |
| POST | `/merchant/recipes` | 新建菜谱规格 |
| PUT | `/merchant/recipes/:recipeId` | 修改菜谱 |
| DELETE | `/merchant/recipes/:recipeId` | 删除菜谱 |

---

#### 菜品管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/merchant/dishes` | 菜品列表 |
| POST | `/merchant/dishes` | 新建菜品 |
| PUT | `/merchant/dishes/:dishId` | 修改菜品 |
| DELETE | `/merchant/dishes/:dishId` | 删除菜品 |

---

#### 分类管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/merchant/categories` | 分类列表 |
| POST | `/merchant/categories` | 新建分类 |
| PUT | `/merchant/categories/:categoryId` | 修改分类 |
| DELETE | `/merchant/categories/:categoryId` | 删除分类 |

---

#### 烹饪任务管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/merchant/tasks` | 任务列表 |
| PUT | `/merchant/tasks/:taskId/reassign` | 重派任务给其他厨师 |

---

#### 服务员管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/merchant/waiters` | 服务员列表 |
| POST | `/merchant/waiters` | 新建服务员 |
| PUT | `/merchant/waiters/:waiterId` | 修改服务员 |
| DELETE | `/merchant/waiters/:waiterId` | 删除服务员 |

---

#### 区域管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/merchant/areas` | 区域列表 |
| POST | `/merchant/areas` | 新建区域 |
| PUT | `/merchant/areas/:areaId` | 修改区域 |
| DELETE | `/merchant/areas/:areaId` | 删除区域 |

---

#### 送餐任务管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/merchant/delivery-tasks` | 送餐任务列表 |
| PUT | `/merchant/delivery-tasks/:taskId/return` | 退回送餐任务 |

---

#### 系统配置

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/merchant/config` | 获取系统配置 |
| PUT | `/merchant/config` | 更新系统配置 |

---

### 7.3 厨师端 `/api/chef`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/chef/auth/login` | 厨师登录 |
| GET | `/chef/chefs/:chefId` | 获取厨师信息 |
| GET | `/chef/tasks` | 获取我的任务列表（需 `?chef_id=`）|
| GET | `/chef/tasks/:taskId` | 任务详情 |
| PUT | `/chef/tasks/:taskId/complete` | 标记任务完成 |

**登录请求体：**
```json
{ "chef_id": 1 }
```

> 厨师登录当前使用 chef_id 简化认证（无密码），生产环境应接入完整鉴权。

**完成任务响应：** 自动将关联的 `order_item` 状态更新为 `done`，并生成 `delivery_task`。

---

### 7.4 服务员端 `/api/waiter`

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/waiter/auth/login` | 服务员登录 |
| GET | `/waiter/waiters/:waiterId` | 获取服务员信息 |
| GET | `/waiter/tasks` | 获取我的送餐任务（需 `?waiter_id=`）|
| GET | `/waiter/tasks/:taskId` | 任务详情 |
| PUT | `/waiter/tasks/:taskId/pickup` | 接单（取餐）|
| PUT | `/waiter/tasks/:taskId/deliver` | 确认送达 |
| PUT | `/waiter/tasks/:taskId/reject` | 拒绝任务 |

**送达后：** 关联订单项状态更新为 `served`，若所有项均已送达则订单更新为 `dining`。

---

## 8. WebSocket 协议

### 8.1 购物车同步 `/ws/cart/:sessionId`

**作用：** 同桌多部手机共享同一购物车会话，任意一端修改购物车（增删改）后，服务端向该 `sessionId` 的所有在线客户端广播完整的 `CartSession` 快照。

**连接地址：**
```
ws://localhost:8080/ws/cart/{sessionId}
```

**消息格式（服务端 → 客户端）：**
```json
{
  "session_id": 5,
  "table_id": 2,
  "status": "active",
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

> 客户端只需监听消息并刷新页面状态，无需主动发送消息。

---

### 8.2 服务员推送 `/ws/waiter/:waiterId`

**作用：** 厨师完成烹饪任务后，系统自动创建送餐任务，并通过 WS 实时推送给对应区域的在线服务员。

**连接地址：**
```
ws://localhost:8080/ws/waiter/{waiterId}
```

**连接时副作用：** 服务员 `status` 字段更新为 `online`；断开连接后自动更新为 `offline`。

**消息格式（服务端 → 客户端）：**
```json
{
  "type": "new_delivery_task",
  "task": {
    "task_id": 8,
    "cooking_task_id": 12,
    "table_ids": [3, 5],
    "status": "pending"
  }
}
```

**区域过滤规则：**
- 服务员 `area_id = 0` 或未设置 → 接收所有区域任务
- 服务员 `area_id = N` → 只接收 `area_id = N` 的任务推送

---

## 9. 自动派发引擎

文件：`backend/services/dispatch.go`

### 9.1 自动派发（RunAutoDispatch）

后台 goroutine 每隔 `DISPATCH_INTERVAL_SEC`（默认 30 秒）触发一次：

```
1. 查询所有 status='pending' 且 created_at >= NOW() - WINDOW_MIN 的 order_item
2. 按 dish_id 分组，汇总：
   - 总份数 (total_qty)
   - 源 item_id 列表 (merged_from)
   - 涉及桌号集合 (table_ids)
   - 是否含 VIP 订单 (has_vip)
3. 对每个菜品调用 createTasksForDish()
```

### 9.2 菜谱选择算法（findBestRecipe）

按优先级选择最优菜谱规格：

```
1. 精确匹配：找 portion == total_qty 的规格
2. 最小覆盖：找最小的 portion >= total_qty 的规格
3. 组合拆分（仅 allow_combine=true）：找最大的 portion <= total_qty 的规格
4. 兜底：使用 portion 最大的规格
```

### 9.3 厨师分配算法（assignChef）

```sql
SELECT * FROM chef
JOIN chef_recipe cr ON cr.chef_id = chef.chef_id AND cr.recipe_id = ?
WHERE chef.is_active = 1
  AND chef.current_load < chef.max_load
ORDER BY chef.current_load ASC
LIMIT 1
```

优先分配**技能匹配 + 当前负载最低**的厨师。

### 9.4 任务优先级

- VIP 订单 → `priority = 10`
- 普通订单 → `priority = 0`

### 9.5 手动派发

- **整单派发**：`POST /merchant/orders/:orderId/dispatch`
  → 对订单内所有 `pending` 的 `order_item` 分别派发
- **单品派发**：`POST /merchant/orders/:orderId/items/:itemId/dispatch`
  → 可指定厨师 `chef_id`，否则自动分配

### 9.6 任务完成后的级联操作

厨师调用 `PUT /chef/tasks/:taskId/complete` 后：

```
1. cooking_task.status = 'done', completed_at = NOW()
2. 关联的 order_item.status = 'done'
3. 厨师 current_load -= task.total_portion
4. 自动创建 delivery_task (status='pending')
5. 通过 WaiterHub 广播新送餐任务给对应区域服务员
6. 若订单所有 item 均为 done/served → order.status = 'dining'
```

---

## 10. 前端应用

### 10.1 顾客端（customer-app）Port 3001

**技术：** Vue3 + Vant 4（移动端 H5）

**路由：**

| 路径 | 组件 | 说明 |
|------|------|------|
| `/order/:tableId` | OrderPage | 点餐主页（扫码后进入）|
| `/status/:orderId` | OrderStatus | 订单实时状态页 |

**核心功能：**
- 扫描二维码携带 `tableId` 参数进入点餐页
- 创建或加入该桌的购物车会话
- 连接 WebSocket，实时接收购物车变化
- 支持多人同时操作，购物车自动同步
- 提交订单后跳转订单状态页，轮询状态更新

---

### 10.2 商家端（merchant-app）Port 3002

**技术：** Vue3 + Vuetify 3（PC 端）

**功能模块：**

| 页面 | 说明 |
|------|------|
| TablesView | 餐桌网格视图，颜色区分桌态；点击查看当前订单；支持增删改桌 |
| OrdersView | 订单列表，支持状态筛选；收款、折扣、手动派发 |
| TasksView | 烹饪任务看板；支持重派厨师 |
| DeliveryTasksView | 送餐任务列表；可退回任务 |
| DishesView | 菜品 CRUD，支持上下架 |
| CategoriesView | 分类 CRUD |
| RecipesView | 菜谱规格 CRUD |
| ChefsView | 厨师 CRUD；管理技能（chef_recipe）|
| WaitersView | 服务员 CRUD；查看在线状态 |
| AreasView | 用餐区域 CRUD |
| ConfigView | 系统参数（派发间隔、时间窗口等）|

**桌态颜色：**

| 状态 | 颜色 |
|------|------|
| idle（空闲）| 绿色 |
| ordering（点餐中）| 蓝色 |
| waiting（等待出餐）| 橙色 |
| dining（用餐中）| 红色 |
| checkout（结账）| 紫色 |

---

### 10.3 厨师端（chef-app）Port 3003

**技术：** Vue3 + Vuetify 3（自适应布局，支持平板/PC）

**页面：**

| 页面 | 说明 |
|------|------|
| LoginView | 厨师选择登录（选择姓名/ID）|
| TasksView | 任务列表：展示菜品名称、规格、合计份数、涉及桌号；一键标记完成 |

---

### 10.4 服务员端（waiter-app）Port 3004

**技术：** Vue3 + Vuetify 3（移动端自适应）

**页面：**

| 页面 | 说明 |
|------|------|
| LoginView | 服务员登录（选择姓名/ID）|
| TasksView | 送餐任务列表；支持接单、确认送达、拒绝；WS 实时推送新任务 |

---

## 11. 配置说明

### 配置加载优先级

```
环境变量 > config.yaml > 代码默认值
```

### 配置文件（`backend/config.yaml`）

```yaml
server:
  port: "8080"

database:
  host: "localhost"
  port: "3306"
  user: "restaurant"
  password: "restaurant123"
  name: "restaurant"

dispatch:
  interval_sec: 30      # 自动派发检查间隔（秒）
  window_min: 5         # 聚合时间窗口（分钟）
```

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `SERVER_PORT` | `8080` | 后端监听端口 |
| `DB_HOST` | `localhost` | 数据库主机 |
| `DB_PORT` | `3306` | 数据库端口 |
| `DB_USER` | `root` | 数据库用户名 |
| `DB_PASSWORD` | `` | 数据库密码 |
| `DB_NAME` | `restaurant` | 数据库名 |
| `DISPATCH_INTERVAL_SEC` | `30` | 自动派发间隔（秒）|
| `DISPATCH_WINDOW_MIN` | `5` | 订单聚合时间窗口（分钟）|
| `CONFIG_FILE` | `config.yaml` | 配置文件路径 |

### 前端环境变量（`.env`）

各前端应用的 `.env.example` 包含：
```
VITE_API_BASE_URL=http://localhost:8080
```

---

## 12. 快速启动

### Docker Compose（推荐）

```bash
# 克隆项目
git clone https://github.com/chengtb/chengtb.git
cd chengtb

# 一键启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f backend
```

**服务访问地址：**

| 服务 | 地址 | 说明 |
|------|------|------|
| 顾客端 | http://localhost:3001 | 扫码点餐（移动端）|
| 商家端 | http://localhost:3002 | 管理后台（PC 端）|
| 厨师端 | http://localhost:3003 | 烹饪任务 |
| 服务员端 | http://localhost:3004 | 送餐任务 |
| 后端 API | http://localhost:8080 | RESTful + WS |
| MySQL | localhost:3306 | restaurant / restaurant123 |

**停止服务：**
```bash
docker-compose down          # 停止并删除容器
docker-compose down -v       # 同时删除数据卷（清空数据库）
```

---

## 13. 本地开发

### 前提条件

- Go 1.21+
- Node.js 18+
- MySQL 8.0（本地或 Docker）

### 启动 MySQL（可用 Docker 单独启动）

```bash
docker run -d \
  --name restaurant-mysql \
  -e MYSQL_DATABASE=restaurant \
  -e MYSQL_USER=restaurant \
  -e MYSQL_PASSWORD=restaurant123 \
  -e MYSQL_ROOT_PASSWORD=rootpassword \
  -p 3306:3306 \
  mysql:8.0
```

### 启动后端

```bash
cd backend

# 方式一：环境变量
export DB_HOST=localhost DB_USER=restaurant DB_PASSWORD=restaurant123 DB_NAME=restaurant
go run main.go

# 方式二：配置文件
cp config.yaml.example config.yaml   # 编辑配置
go run main.go
```

### 启动前端应用

```bash
# 顾客端
cd customer-app
cp .env.example .env
npm install
npm run dev       # http://localhost:5173

# 商家端
cd merchant-app
cp .env.example .env
npm install
npm run dev       # http://localhost:5174

# 厨师端
cd chef-app
cp .env.example .env
npm install
npm run dev       # http://localhost:5175

# 服务员端
cd waiter-app
cp .env.example .env
npm install
npm run dev       # http://localhost:5176
```

### 构建生产版本

```bash
cd customer-app && npm run build    # 输出 dist/
cd merchant-app && npm run build
cd chef-app && npm run build
cd waiter-app && npm run build
```

---

## 14. 常见问题

### Q1：多人扫码点餐时购物车不同步？

**A：** 检查前端 WebSocket 连接是否成功建立。确保所有客户端连接到相同的 `sessionId`（同一张桌的 `cart_session.session_id`）。

### Q2：自动派发没有创建任务？

**A：** 常见原因：
1. 该菜品没有对应的 `recipe` 记录，或 `is_enabled=false`
2. 没有技能匹配的活跃厨师（`chef_recipe` 表未配置）
3. 所有符合技能的厨师 `current_load >= max_load`
4. 订单创建时间超出了 `DISPATCH_WINDOW_MIN` 时间窗口

### Q3：GORM 关联查询结果为空？

**A：** 已知问题：在 GORM v1.30.0 中，对结构体指针关联使用 `gorm:"foreignKey:XxxID"` 标签时，若 `XxxID` 同时是被关联模型的主键，GORM 会误判为 `has-one` 关系。**解决方案：** 删除该标签（让 GORM 自动推断）或补充 `references:XxxID`。参见 `backend/models/recipe.go`。

### Q4：数据库迁移失败？

**A：** `migrations.go` 中的 `ALTER TABLE ... ADD COLUMN IF NOT EXISTS` 仅 MySQL 8.0+ 支持。确保 MySQL 版本 >= 8.0。

### Q5：CORS 跨域错误？

**A：** 当前 `middleware/cors.go` 允许所有来源（开发用）。生产部署时请修改为指定的前端域名。

### Q6：如何配置厨师技能？

**A：** 在商家端 → 厨师管理 → 编辑厨师 → 选择可烹饪的菜谱规格（`chef_recipe` 表）。未配置技能的厨师不会收到任何任务。

---

*文档生成时间：2026-03-30*
