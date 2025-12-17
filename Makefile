.PHONY: help build run clean docker-up docker-down test

help:
	@echo "Golang Nacos+Nats 微服务示例"
	@echo ""
	@echo "可用命令:"
	@echo "  make docker-up    - 启动 Nacos 和 NATS 基础设施"
	@echo "  make docker-down  - 停止基础设施"
	@echo "  make build        - 编译所有微服务"
	@echo "  make run-producer - 运行生产者服务"
	@echo "  make run-consumer - 运行消费者服务"
	@echo "  make run-gateway  - 运行 API 网关"
	@echo "  make clean        - 清理编译文件"
	@echo "  make deps         - 下载依赖"

# 启动基础设施
docker-up:
	@echo "启动 Nacos 和 NATS..."
	docker-compose up -d
	@echo "等待服务就绪..."
	@sleep 10
	@echo "服务已启动！"
	@echo "Nacos 控制台: http://localhost:8848/nacos (nacos/nacos)"
	@echo "NATS 监控: http://localhost:8222"

# 停止基础设施
docker-down:
	@echo "停止基础设施..."
	docker-compose down

# 下载依赖
deps:
	@echo "下载依赖..."
	cd producer && go mod download
	cd consumer && go mod download
	cd gateway && go mod download

# 编译所有服务
build:
	@echo "编译所有服务..."
	cd producer && go build -o producer main.go
	cd consumer && go build -o consumer main.go
	cd gateway && go build -o gateway main.go
	@echo "编译完成！"

# 运行生产者服务
run-producer:
	@echo "运行生产者服务..."
	cd producer && go run main.go

# 运行消费者服务
run-consumer:
	@echo "运行消费者服务..."
	cd consumer && go run main.go

# 运行 API 网关
run-gateway:
	@echo "运行 API 网关..."
	cd gateway && go run main.go

# 清理编译文件
clean:
	@echo "清理编译文件..."
	rm -f producer/producer
	rm -f consumer/consumer
	rm -f gateway/gateway
	rm -rf /tmp/nacos
	@echo "清理完成！"

# 测试 API
test-api:
	@echo "测试服务列表..."
	curl http://localhost:8080/api/services
	@echo "\n测试健康检查..."
	curl http://localhost:8080/health
	@echo "\n测试发布消息..."
	curl -X POST http://localhost:8080/api/publish \
		-H "Content-Type: application/json" \
		-d '{"content": "测试消息"}'
