# 餐厅点餐与厨房管理系统 API 文档

## 目录

- [概述](#概述)
- [通用说明](#通用说明)
- [商品与菜单](#商品与菜单)
- [餐桌与区域](#餐桌与区域)
- [订单管理](#订单管理)
- [厨房管理](#厨房管理)
- [退款管理](#退款管理)
- [赠菜管理](#赠菜管理)
- [WebSocket 协议](#websocket-协议)
- [状态码与错误处理](#状态码与错误处理)

---

## 概述

本文档描述餐厅点餐与厨房管理系统的 RESTful API 和 WebSocket 实时通信协议。

- **Base URL:** `http://<host>:<port>/api/v1/`
- **数据格式:** JSON
- **字符编码:** UTF-8

---

## 通用说明

### 请求头

```
Content-Type: application/json
```

### 通用响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

### 分页参数

| 参数     | 类型   | 必填 | 说明               |
| -------- | ------ | ---- | ------------------ |
| `page`   | int    | 否   | 页码，默认 1       |
| `limit`  | int    | 否   | 每页条数，默认 20  |

### 分页响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [],
    "total": 100,
    "page": 1,
    "limit": 20
  }
}
```

---

## 商品与菜单

### 获取分类列表（含商品）

**GET** `/api/v1/categories`

获取所有菜品分类及其下属商品列表。

**请求示例：**

```
GET /api/v1/categories
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "热菜",
      "sort": 1,
      "products": [
        {
          "id": 1,
          "name": "宫保鸡丁",
          "price": 38.00,
          "image": "/uploads/gbjd.jpg",
          "description": "经典川菜",
          "category_id": 1,
          "status": 1
        }
      ]
    }
  ]
}
```

---

### 获取商品列表

**GET** `/api/v1/products`

获取商品列表，可按分类筛选。

**查询参数：**

| 参数          | 类型   | 必填 | 说明           |
| ------------- | ------ | ---- | -------------- |
| `category_id` | int    | 否   | 按分类 ID 筛选 |

**请求示例：**

```
GET /api/v1/products?category_id=1
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "宫保鸡丁",
      "price": 38.00,
      "image": "/uploads/gbjd.jpg",
      "description": "经典川菜",
      "category_id": 1,
      "status": 1
    }
  ]
}
```

---

### 获取商品详情（含规格）

**GET** `/api/v1/products/:id`

获取单个商品的详细信息，包括规格选项。

**路径参数：**

| 参数 | 类型 | 说明    |
| ---- | ---- | ------- |
| `id` | int  | 商品 ID |

**请求示例：**

```
GET /api/v1/products/1
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "name": "宫保鸡丁",
    "price": 38.00,
    "image": "/uploads/gbjd.jpg",
    "description": "经典川菜",
    "category_id": 1,
    "status": 1,
    "specs": [
      {
        "id": 1,
        "product_id": 1,
        "name": "微辣",
        "price_adjustment": 0
      },
      {
        "id": 2,
        "product_id": 1,
        "name": "大份",
        "price_adjustment": 10.00
      }
    ]
  }
}
```

---

## 餐桌与区域

### 获取就餐区域列表（含餐桌）

**GET** `/api/v1/regions`

获取所有就餐区域及其包含的餐桌信息。

**请求示例：**

```
GET /api/v1/regions
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "大厅",
      "tables": [
        {
          "id": 1,
          "region_id": 1,
          "name": "A1",
          "capacity": 4,
          "status": "idle"
        },
        {
          "id": 2,
          "region_id": 1,
          "name": "A2",
          "capacity": 6,
          "status": "occupied"
        }
      ]
    }
  ]
}
```

---

### 获取餐桌信息

**GET** `/api/v1/tables/:id`

获取单个餐桌的详细信息。

**路径参数：**

| 参数 | 类型 | 说明    |
| ---- | ---- | ------- |
| `id` | int  | 餐桌 ID |

**请求示例：**

```
GET /api/v1/tables/1
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "region_id": 1,
    "name": "A1",
    "capacity": 4,
    "status": "idle"
  }
}
```

---

## 订单管理

### 创建订单（顾客提交）

**POST** `/api/v1/orders`

顾客在客户端提交点餐订单。

**请求体：**

```json
{
  "table_id": 1,
  "items": [
    {
      "product_id": 1,
      "spec_id": 1,
      "quantity": 2,
      "remark": "少放辣"
    },
    {
      "product_id": 3,
      "spec_id": null,
      "quantity": 1,
      "remark": ""
    }
  ],
  "remark": "请尽快上菜"
}
```

**响应示例：**

```json
{
  "code": 200,
  "message": "订单创建成功",
  "data": {
    "id": 101,
    "order_no": "ORD202401010001",
    "table_id": 1,
    "status": "pending",
    "total_amount": 86.00,
    "remark": "请尽快上菜",
    "items": [
      {
        "id": 1,
        "product_id": 1,
        "product_name": "宫保鸡丁",
        "spec_name": "微辣",
        "quantity": 2,
        "price": 38.00,
        "remark": "少放辣"
      }
    ],
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

---

### 获取订单列表

**GET** `/api/v1/orders`

获取订单列表，支持状态筛选和分页。

**查询参数：**

| 参数     | 类型   | 必填 | 说明                                              |
| -------- | ------ | ---- | ------------------------------------------------- |
| `status` | string | 否   | 订单状态：`pending`/`confirmed`/`completed`/`cancelled` |
| `page`   | int    | 否   | 页码，默认 1                                      |
| `limit`  | int    | 否   | 每页条数，默认 20                                  |

**请求示例：**

```
GET /api/v1/orders?status=pending&page=1&limit=10
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 101,
        "order_no": "ORD202401010001",
        "table_id": 1,
        "table_name": "A1",
        "status": "pending",
        "total_amount": 86.00,
        "created_at": "2024-01-01T12:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "limit": 10
  }
}
```

---

### 获取订单详情

**GET** `/api/v1/orders/:id`

获取单个订单的完整详情。

**路径参数：**

| 参数 | 类型 | 说明    |
| ---- | ---- | ------- |
| `id` | int  | 订单 ID |

**请求示例：**

```
GET /api/v1/orders/101
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 101,
    "order_no": "ORD202401010001",
    "table_id": 1,
    "table_name": "A1",
    "status": "pending",
    "total_amount": 86.00,
    "remark": "请尽快上菜",
    "items": [
      {
        "id": 1,
        "product_id": 1,
        "product_name": "宫保鸡丁",
        "spec_name": "微辣",
        "quantity": 2,
        "price": 38.00,
        "remark": "少放辣"
      }
    ],
    "logs": [
      {
        "action": "created",
        "operator": "顾客",
        "created_at": "2024-01-01T12:00:00Z"
      }
    ],
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

---

### 确认订单（服务员）

**PUT** `/api/v1/orders/:id/confirm`

服务员确认顾客提交的订单。

**路径参数：**

| 参数 | 类型 | 说明    |
| ---- | ---- | ------- |
| `id` | int  | 订单 ID |

**请求示例：**

```
PUT /api/v1/orders/101/confirm
```

**响应示例：**

```json
{
  "code": 200,
  "message": "订单已确认",
  "data": {
    "id": 101,
    "status": "confirmed"
  }
}
```

---

### 修改订单项（服务员）

**PUT** `/api/v1/orders/:id/items`

服务员修改订单中的菜品项。

**路径参数：**

| 参数 | 类型 | 说明    |
| ---- | ---- | ------- |
| `id` | int  | 订单 ID |

**请求体：**

```json
{
  "items": [
    {
      "product_id": 1,
      "spec_id": 1,
      "quantity": 3,
      "remark": "加辣"
    },
    {
      "product_id": 5,
      "spec_id": null,
      "quantity": 1,
      "remark": ""
    }
  ]
}
```

**响应示例：**

```json
{
  "code": 200,
  "message": "订单已更新",
  "data": {
    "id": 101,
    "total_amount": 124.00,
    "items": [
      {
        "id": 1,
        "product_id": 1,
        "product_name": "宫保鸡丁",
        "spec_name": "微辣",
        "quantity": 3,
        "price": 38.00,
        "remark": "加辣"
      },
      {
        "id": 4,
        "product_id": 5,
        "product_name": "蒜蓉西兰花",
        "spec_name": null,
        "quantity": 1,
        "price": 22.00,
        "remark": ""
      }
    ]
  }
}
```

---

### 获取餐桌订单

**GET** `/api/v1/tables/:table_id/orders`

获取指定餐桌的所有订单。

**路径参数：**

| 参数       | 类型 | 说明    |
| ---------- | ---- | ------- |
| `table_id` | int  | 餐桌 ID |

**请求示例：**

```
GET /api/v1/tables/1/orders
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 101,
      "order_no": "ORD202401010001",
      "table_id": 1,
      "status": "confirmed",
      "total_amount": 86.00,
      "created_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

---

## 厨房管理

### 拆分订单为厨房任务

**POST** `/api/v1/orders/:id/split`

将已确认的订单拆分为厨房任务，分配给不同的厨师。

**路径参数：**

| 参数 | 类型 | 说明    |
| ---- | ---- | ------- |
| `id` | int  | 订单 ID |

**请求示例：**

```
POST /api/v1/orders/101/split
```

**响应示例：**

```json
{
  "code": 200,
  "message": "任务拆分成功",
  "data": [
    {
      "id": 1,
      "order_id": 101,
      "chef_id": 1,
      "chef_name": "张师傅",
      "status": "pending",
      "items": [
        {
          "id": 1,
          "product_name": "宫保鸡丁",
          "quantity": 2,
          "remark": "少放辣"
        }
      ],
      "created_at": "2024-01-01T12:05:00Z"
    }
  ]
}
```

---

### 获取厨房任务列表

**GET** `/api/v1/kitchen/tasks`

获取厨房任务列表，供厨师和 KDS 使用。

**查询参数：**

| 参数      | 类型   | 必填 | 说明                                     |
| --------- | ------ | ---- | ---------------------------------------- |
| `status`  | string | 否   | 任务状态：`pending`/`cooking`/`completed` |
| `chef_id` | int    | 否   | 按厨师 ID 筛选                           |

**请求示例：**

```
GET /api/v1/kitchen/tasks?status=pending
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "order_id": 101,
      "order_no": "ORD202401010001",
      "table_name": "A1",
      "chef_id": 1,
      "chef_name": "张师傅",
      "status": "pending",
      "items": [
        {
          "id": 1,
          "product_name": "宫保鸡丁",
          "quantity": 2,
          "remark": "少放辣"
        }
      ],
      "created_at": "2024-01-01T12:05:00Z"
    }
  ]
}
```

---

### 完成厨房任务

**PUT** `/api/v1/kitchen/tasks/:id/complete`

厨师标记厨房任务为已完成（菜品已出餐）。

**路径参数：**

| 参数 | 类型 | 说明    |
| ---- | ---- | ------- |
| `id` | int  | 任务 ID |

**请求示例：**

```
PUT /api/v1/kitchen/tasks/1/complete
```

**响应示例：**

```json
{
  "code": 200,
  "message": "任务已完成",
  "data": {
    "id": 1,
    "status": "completed",
    "completed_at": "2024-01-01T12:20:00Z"
  }
}
```

---

### 重新分配厨房任务

**PUT** `/api/v1/kitchen/tasks/:id/reassign`

将厨房任务重新分配给其他厨师。

**路径参数：**

| 参数 | 类型 | 说明    |
| ---- | ---- | ------- |
| `id` | int  | 任务 ID |

**请求体：**

```json
{
  "chef_id": 2
}
```

**响应示例：**

```json
{
  "code": 200,
  "message": "任务已重新分配",
  "data": {
    "id": 1,
    "chef_id": 2,
    "chef_name": "李师傅",
    "status": "pending"
  }
}
```

---

## 退款管理

### 创建退款

**POST** `/api/v1/refunds`

创建退款申请。

**请求体：**

```json
{
  "order_id": 101,
  "order_item_id": 1,
  "reason": "菜品口味不符",
  "amount": 38.00
}
```

**响应示例：**

```json
{
  "code": 200,
  "message": "退款创建成功",
  "data": {
    "id": 1,
    "order_id": 101,
    "order_item_id": 1,
    "reason": "菜品口味不符",
    "amount": 38.00,
    "status": "pending",
    "created_at": "2024-01-01T13:00:00Z"
  }
}
```

---

### 获取退款列表

**GET** `/api/v1/refunds`

获取退款记录列表。

**请求示例：**

```
GET /api/v1/refunds
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "order_id": 101,
      "order_item_id": 1,
      "reason": "菜品口味不符",
      "amount": 38.00,
      "status": "approved",
      "created_at": "2024-01-01T13:00:00Z"
    }
  ]
}
```

---

## 赠菜管理

### 创建赠菜

**POST** `/api/v1/gifts`

创建赠菜记录。

**请求体：**

```json
{
  "order_id": 101,
  "product_id": 8,
  "quantity": 1,
  "reason": "老顾客回馈"
}
```

**响应示例：**

```json
{
  "code": 200,
  "message": "赠菜创建成功",
  "data": {
    "id": 1,
    "order_id": 101,
    "product_id": 8,
    "product_name": "水果拼盘",
    "quantity": 1,
    "reason": "老顾客回馈",
    "created_at": "2024-01-01T13:10:00Z"
  }
}
```

---

### 获取赠菜列表

**GET** `/api/v1/gifts`

获取赠菜记录列表。

**请求示例：**

```
GET /api/v1/gifts
```

**响应示例：**

```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "id": 1,
      "order_id": 101,
      "product_id": 8,
      "product_name": "水果拼盘",
      "quantity": 1,
      "reason": "老顾客回馈",
      "created_at": "2024-01-01T13:10:00Z"
    }
  ]
}
```

---

## WebSocket 协议

### 连接

**URL:** `ws://<host>:<port>/ws?table_id=<TABLE_ID>&type=<CLIENT_TYPE>`

**连接参数：**

| 参数       | 类型   | 必填 | 说明                                           |
| ---------- | ------ | ---- | ---------------------------------------------- |
| `table_id` | int    | 是   | 餐桌 ID                                       |
| `type`     | string | 是   | 客户端类型：`customer`/`waiter`/`chef`/`kds`   |

**客户端类型说明：**

| 类型       | 说明                               |
| ---------- | ---------------------------------- |
| `customer` | 顾客端，用于点餐和接收订单状态     |
| `waiter`   | 服务员端，用于接收新订单和出餐通知 |
| `chef`     | 厨师端，用于接收和管理厨房任务     |
| `kds`      | 厨房显示系统，用于展示所有厨房任务 |

**连接示例：**

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?table_id=1&type=customer');

ws.onopen = function() {
  console.log('WebSocket 连接已建立');
};

ws.onmessage = function(event) {
  const message = JSON.parse(event.data);
  console.log('收到消息:', message);
};

ws.onclose = function() {
  console.log('WebSocket 连接已关闭');
};
```

---

### 消息格式

所有 WebSocket 消息采用 JSON 格式：

```json
{
  "type": "<消息类型>",
  "data": {}
}
```

---

### 消息类型

#### `cart_update` — 购物车同步

同一餐桌的顾客之间同步购物车内容。

**方向：** 顾客 → 服务器 → 同桌其他顾客

```json
{
  "type": "cart_update",
  "data": {
    "table_id": 1,
    "items": [
      {
        "product_id": 1,
        "product_name": "宫保鸡丁",
        "spec_id": 1,
        "spec_name": "微辣",
        "quantity": 2,
        "price": 38.00,
        "remark": "少放辣"
      }
    ],
    "total_amount": 76.00
  }
}
```

---

#### `new_order` — 新订单通知

通知服务员有新订单需要处理。

**方向：** 服务器 → 服务员

```json
{
  "type": "new_order",
  "data": {
    "order_id": 101,
    "order_no": "ORD202401010001",
    "table_id": 1,
    "table_name": "A1",
    "total_amount": 86.00,
    "item_count": 3,
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

---

#### `order_created` — 订单确认通知

通知顾客订单已成功创建。

**方向：** 服务器 → 顾客

```json
{
  "type": "order_created",
  "data": {
    "order_id": 101,
    "order_no": "ORD202401010001",
    "status": "pending",
    "message": "订单已提交，等待服务员确认"
  }
}
```

---

#### `order_confirmed` — 订单已确认通知

通知顾客订单已被服务员确认。

**方向：** 服务器 → 顾客

```json
{
  "type": "order_confirmed",
  "data": {
    "order_id": 101,
    "order_no": "ORD202401010001",
    "status": "confirmed",
    "message": "订单已确认，正在准备中"
  }
}
```

---

#### `new_tasks` — 新厨房任务通知

通知厨师和 KDS 有新的厨房任务。

**方向：** 服务器 → 厨师/KDS

```json
{
  "type": "new_tasks",
  "data": {
    "tasks": [
      {
        "id": 1,
        "order_id": 101,
        "order_no": "ORD202401010001",
        "table_name": "A1",
        "chef_id": 1,
        "chef_name": "张师傅",
        "items": [
          {
            "product_name": "宫保鸡丁",
            "quantity": 2,
            "remark": "少放辣"
          }
        ]
      }
    ]
  }
}
```

---

#### `task_assigned` — 任务分配通知

通知厨师有新任务被分配。

**方向：** 服务器 → 厨师

```json
{
  "type": "task_assigned",
  "data": {
    "task_id": 1,
    "order_no": "ORD202401010001",
    "table_name": "A1",
    "chef_id": 1,
    "items": [
      {
        "product_name": "宫保鸡丁",
        "quantity": 2,
        "remark": "少放辣"
      }
    ]
  }
}
```

---

#### `task_reassigned` — 任务重新分配通知

通知相关厨师任务已被重新分配。

**方向：** 服务器 → 厨师

```json
{
  "type": "task_reassigned",
  "data": {
    "task_id": 1,
    "order_no": "ORD202401010001",
    "from_chef_id": 1,
    "from_chef_name": "张师傅",
    "to_chef_id": 2,
    "to_chef_name": "李师傅"
  }
}
```

---

#### `dish_ready` — 菜品出餐通知

通知服务员菜品已准备完成，可以上菜。

**方向：** 服务器 → 服务员

```json
{
  "type": "dish_ready",
  "data": {
    "task_id": 1,
    "order_id": 101,
    "order_no": "ORD202401010001",
    "table_id": 1,
    "table_name": "A1",
    "items": [
      {
        "product_name": "宫保鸡丁",
        "quantity": 2
      }
    ],
    "chef_name": "张师傅",
    "completed_at": "2024-01-01T12:20:00Z"
  }
}
```

---

## 状态码与错误处理

### HTTP 状态码

| 状态码 | 说明                 |
| ------ | -------------------- |
| 200    | 请求成功             |
| 400    | 请求参数错误         |
| 404    | 资源不存在           |
| 409    | 状态冲突（如重复操作）|
| 500    | 服务器内部错误       |

### 业务错误码

| code  | 说明                     |
| ----- | ------------------------ |
| 200   | 成功                     |
| 40001 | 参数校验失败             |
| 40002 | 商品不存在或已下架       |
| 40003 | 餐桌不存在               |
| 40004 | 订单不存在               |
| 40005 | 订单状态不允许此操作     |
| 40006 | 厨房任务不存在           |
| 40007 | 厨师不存在               |
| 40008 | 退款金额超出订单金额     |
| 50001 | 数据库操作失败           |
| 50002 | WebSocket 连接失败       |

### 错误响应示例

```json
{
  "code": 40004,
  "message": "订单不存在",
  "data": null
}
```

```json
{
  "code": 40005,
  "message": "当前订单状态不允许此操作",
  "data": {
    "current_status": "completed",
    "allowed_statuses": ["pending"]
  }
}
```

### 订单状态流转

```
pending（待确认）→ confirmed（已确认）→ completed（已完成）
                                      → cancelled（已取消）
```

### 厨房任务状态流转

```
pending（待处理）→ cooking（制作中）→ completed（已完成）
```
