package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/nats-io/nats.go"
)

const (
	ServiceName = "consumer-service"
	ServicePort = 8082
	NatsSubject = "microservice.events"
)

type Consumer struct {
	natsConn    *nats.Conn
	nacosClient naming_client.INamingClient
	sub         *nats.Subscription
	serviceIP   string
	servicePort uint64
}

func NewConsumer(natsURL, nacosAddr string, nacosPort uint64) (*Consumer, error) {
	// 连接 NATS
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// 配置 Nacos 客户端
	serverConfigs := []constant.ServerConfig{
		*constant.NewServerConfig(nacosAddr, nacosPort, constant.WithContextPath("/nacos")),
	}

	clientConfig := constant.ClientConfig{
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "info",
	}

	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create Nacos client: %w", err)
	}

	return &Consumer{
		natsConn:    nc,
		nacosClient: namingClient,
	}, nil
}

func (c *Consumer) RegisterService(ip string, port uint64) error {
	success, err := c.nacosClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          ip,
		Port:        port,
		ServiceName: ServiceName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    map[string]string{"version": "1.0"},
	})

	if err != nil || !success {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// 保存注册信息用于注销
	c.serviceIP = ip
	c.servicePort = port

	log.Printf("Service %s registered successfully at %s:%d", ServiceName, ip, port)
	return nil
}

func (c *Consumer) Subscribe() error {
	sub, err := c.natsConn.Subscribe(NatsSubject, func(msg *nats.Msg) {
		log.Printf("Received message: %s", string(msg.Data))
	})
	
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	c.sub = sub
	log.Printf("Subscribed to subject: %s", NatsSubject)
	return nil
}

func (c *Consumer) Close() {
	if c.sub != nil {
		c.sub.Unsubscribe()
	}
	if c.natsConn != nil {
		c.natsConn.Close()
	}
	if c.nacosClient != nil && c.serviceIP != "" {
		c.nacosClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          c.serviceIP,
			Port:        c.servicePort,
			ServiceName: ServiceName,
		})
	}
}

func main() {
	log.Println("Starting Consumer Service...")

	consumer, err := NewConsumer(
		"nats://localhost:4222",
		"127.0.0.1",
		8848,
	)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	// 注册服务到 Nacos
	if err := consumer.RegisterService("127.0.0.1", ServicePort); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 订阅消息
	if err := consumer.Subscribe(); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	log.Println("Consumer service is running... Press Ctrl+C to exit")

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down consumer service...")
}
