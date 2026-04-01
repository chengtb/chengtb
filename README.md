# 餐厅扫码点餐与后厨管理系统（KDS）

一套完整的餐厅数字化解决方案，包含顾客点餐、服务员管理、后厨调度等功能。

## 技术栈

- **后端**：Go + Gin + GORM + MySQL + WebSocket
- **顾客端**：Vue3 + Vant（移动端）
- **服务员端**：Vue3 + Vuetify（平板/PC）
- **后厨端**：Vue3 + Vuetify（触摸屏优化，暗色主题）

## 项目结构

```
├── backend/                 # Go 后端服务
│   ├── cmd/main.go         # 入口文件
│   └── internal/
│       ├── config/         # 配置管理
│       ├── handlers/       # API 处理器
│       ├── middleware/     # 中间件（CORS、响应封装）
│       ├── models/         # GORM 数据模型（18张表）
│       ├── services/       # 业务逻辑层
│       └── websocket/      # WebSocket 实时通信
├── frontend/
│   ├── customer-app/       # 顾客点餐端（Vue3 + Vant）
│   ├── waiter-app/         # 服务员端（Vue3 + Vuetify）
│   └── kds-app/            # 后厨管理端（Vue3 + Vuetify）
└── docs/
    └── API.md              # API 接口文档
```

## 核心功能

1. **顾客点餐**：扫码点餐、分类浏览、规格选择、购物车实时同步（WebSocket）
2. **服务员确认**：实时订单通知、增删改菜、确认出单
3. **后厨调度**：订单自动拆解、菜谱匹配、智能合单、厨师分配
4. **退菜管理**：三级权限控制（服务员/领班/店长）
5. **赠送管理**：补偿性赠送、客户关系维护、营销活动三种场景
6. **优惠系统**：优惠券、活动促销

## 快速开始

### 环境要求
- Go 1.20+
- Node.js 18+
- MySQL 8.0+

### 后端启动
```bash
cd backend
export DB_HOST=127.0.0.1 DB_PORT=3306 DB_USER=root DB_PASSWORD=yourpass DB_NAME=restaurant_kds
go mod tidy
go run cmd/main.go
```

### 前端启动（开发模式）
```bash
# 顾客端
cd frontend/customer-app && npm install && npm run dev

# 服务员端
cd frontend/waiter-app && npm install && npm run dev

# 后厨端
cd frontend/kds-app && npm install && npm run dev
```

### 运行测试
```bash
cd backend && CGO_ENABLED=1 go test ./internal/services/ -v
```

## API 文档

详见 [docs/API.md](docs/API.md)

## WebSocket 协议

连接地址：`ws://host/ws?table_id=X&type=customer|waiter|chef|kds`

| 消息类型 | 说明 | 接收方 |
|---------|------|--------|
| `cart_update` | 购物车同步 | 同桌顾客 |
| `new_order` | 新订单通知 | 服务员 |
| `order_confirmed` | 订单已确认 | 顾客 |
| `new_tasks` | 新任务 | KDS/厨师 |
| `dish_ready` | 菜品完成 | 服务员 |
