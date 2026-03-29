# 餐厅扫码点餐系统

A complete restaurant QR code ordering system with Go backend and Vue 3 frontends.

## System Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Mobile App  │    │  Admin App  │    │ Kitchen App  │
│  Port: 5173  │    │  Port: 5174  │    │  Port: 5175  │
└──────┬──────┘    └──────┬──────┘    └──────┬──────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │ REST API
                  ┌────────┴────────┐
                  │   Go Backend    │
                  │   Port: 8080    │
                  └─────────────────┘
```

## Components

- **server/** - Go + Gin + GORM + SQLite REST API
- **mobile/** - Vue 3 customer ordering app (accessed via QR code)
- **admin/** - Vue 3 merchant management dashboard
- **kitchen/** - Vue 3 kitchen display system

## Quick Start

### Start the backend:
```bash
cd server && go run main.go
```

### Start the mobile app:
```bash
cd mobile && npm run dev
```

### Start the admin app:
```bash
cd admin && npm run dev
```

### Start the kitchen app:
```bash
cd kitchen && npm run dev
```

## Usage

1. Open Admin app at http://localhost:5174
2. Go to 桌台总览 (Tables Overview)
3. Click a table and download its QR code
4. Scan the QR code with a phone to open the mobile ordering app
5. Order food - it appears in the kitchen app at http://localhost:5175
6. Kitchen staff updates cooking status
7. Admin can manage orders, apply discounts, and mark as paid

## API Reference

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/tables | List all tables |
| GET | /api/tables/:id | Get table info |
| PUT | /api/tables/:id/status | Update table status |
| GET | /api/categories | List categories |
| POST | /api/categories | Create category |
| PUT | /api/categories/:id | Update category |
| DELETE | /api/categories/:id | Delete category |
| GET | /api/dishes | List dishes |
| POST | /api/dishes | Create dish |
| PUT | /api/dishes/:id | Update dish |
| DELETE | /api/dishes/:id | Delete dish |
| POST | /api/orders | Create order |
| GET | /api/orders | List orders |
| GET | /api/orders/:id | Get order details |
| PUT | /api/orders/:id/status | Update order status |
| POST | /api/orders/:id/items | Add items to order |
| GET | /api/cooking-tasks | List cooking tasks |
| PUT | /api/cooking-tasks/:id/status | Update cooking task |
| GET | /api/qrcode/:tableId | Get table QR code |
