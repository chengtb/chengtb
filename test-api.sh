#!/bin/bash

# API 测试脚本
# 确保 Gateway 服务正在运行在 http://localhost:8080

echo "========================================="
echo "测试 Golang Nacos+Nats 微服务 API"
echo "========================================="
echo ""

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查 Gateway 是否运行
echo -e "${BLUE}1. 检查 Gateway 健康状态...${NC}"
HEALTH=$(curl -s http://localhost:8080/health)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Gateway 运行正常${NC}"
    echo "   响应: $HEALTH"
else
    echo -e "${RED}✗ Gateway 未运行或不可访问${NC}"
    echo "   请先启动 Gateway: cd gateway && go run main.go"
    exit 1
fi
echo ""

# 查询已注册的服务
echo -e "${BLUE}2. 查询所有已注册的服务...${NC}"
SERVICES=$(curl -s http://localhost:8080/api/services)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 成功获取服务列表${NC}"
    echo "$SERVICES" | jq '.' 2>/dev/null || echo "$SERVICES"
else
    echo -e "${RED}✗ 获取服务列表失败${NC}"
fi
echo ""

# 发布测试消息 1
echo -e "${BLUE}3. 发布测试消息 (测试 1)...${NC}"
RESPONSE=$(curl -s -X POST http://localhost:8080/api/publish \
  -H "Content-Type: application/json" \
  -d '{"content": "测试消息 #1: Hello from test script!"}')
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 消息发布成功${NC}"
    echo "   响应: $RESPONSE"
else
    echo -e "${RED}✗ 消息发布失败${NC}"
fi
echo ""

# 等待一下再发送下一条消息
sleep 1

# 发布测试消息 2
echo -e "${BLUE}4. 发布测试消息 (测试 2)...${NC}"
RESPONSE=$(curl -s -X POST http://localhost:8080/api/publish \
  -H "Content-Type: application/json" \
  -d '{"content": "测试消息 #2: Nacos + NATS 集成测试"}')
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 消息发布成功${NC}"
    echo "   响应: $RESPONSE"
else
    echo -e "${RED}✗ 消息发布失败${NC}"
fi
echo ""

# 发布包含中文的消息
echo -e "${BLUE}5. 发布中文消息 (测试 3)...${NC}"
RESPONSE=$(curl -s -X POST http://localhost:8080/api/publish \
  -H "Content-Type: application/json" \
  -d '{"content": "测试消息 #3: 这是一条包含中文的消息 🚀"}')
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 中文消息发布成功${NC}"
    echo "   响应: $RESPONSE"
else
    echo -e "${RED}✗ 中文消息发布失败${NC}"
fi
echo ""

# 测试总结
echo "========================================="
echo -e "${GREEN}测试完成！${NC}"
echo "========================================="
echo ""
echo "请检查 Consumer 服务的终端输出，应该能看到以上发送的所有消息。"
echo ""
echo "访问 Nacos 控制台查看服务: http://localhost:8848/nacos"
echo "访问 NATS 监控页面: http://localhost:8222"
echo ""
