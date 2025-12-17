#!/bin/bash

echo "========================================"
echo "启动 Golang Nacos+Nats 微服务示例"
echo "========================================"
echo ""

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo "错误: Docker 未运行，请先启动 Docker"
    exit 1
fi

# 启动基础设施
echo "1. 启动 Nacos 和 NATS 基础设施..."
docker-compose up -d

echo "   等待服务就绪..."
sleep 15

# 检查服务状态
echo ""
echo "2. 检查服务状态..."
docker-compose ps

echo ""
echo "3. 检查 Nacos 健康状态..."
curl -s http://localhost:8848/nacos/v1/console/health/liveness || echo "Nacos 未就绪"

echo ""
echo "4. 检查 NATS 状态..."
curl -s http://localhost:8222/healthz || echo "NATS 未就绪"

echo ""
echo "========================================"
echo "基础设施已启动！"
echo "========================================"
echo ""
echo "访问地址："
echo "  - Nacos 控制台: http://localhost:8848/nacos (用户名/密码: nacos/nacos)"
echo "  - NATS 监控: http://localhost:8222"
echo ""
echo "下一步："
echo "  在不同的终端窗口中运行以下命令启动微服务："
echo ""
echo "  终端 1: cd consumer && go run main.go"
echo "  终端 2: cd producer && go run main.go"
echo "  终端 3: cd gateway && go run main.go"
echo ""
echo "或者使用 Makefile:"
echo "  make run-consumer  # 在终端 1"
echo "  make run-producer  # 在终端 2"
echo "  make run-gateway   # 在终端 3"
echo ""
echo "测试 API："
echo "  make test-api"
echo ""
echo "========================================"
