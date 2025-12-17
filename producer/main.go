package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/nats-io/nats.go"
)

const (
	ServiceName = "producer-service"
	ServicePort = 8081
	NatsSubject = "microservice.events"
)

type Producer struct {
	natsConn    *nats.Conn
	nacosClient naming_client.INamingClient
	serviceIP   string
	servicePort uint64
}

func NewProducer(natsURL, nacosAddr string, nacosPort uint64) (*Producer, error) {
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

	return &Producer{
		natsConn:    nc,
		nacosClient: namingClient,
	}, nil
}

func (p *Producer) RegisterService(ip string, port uint64) error {
	success, err := p.nacosClient.RegisterInstance(vo.RegisterInstanceParam{
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
	p.serviceIP = ip
	p.servicePort = port

	log.Printf("Service %s registered successfully at %s:%d", ServiceName, ip, port)
	return nil
}

func (p *Producer) PublishMessage(message string) error {
	return p.natsConn.Publish(NatsSubject, []byte(message))
}

func (p *Producer) Start() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	counter := 0
	for range ticker.C {
		counter++
		message := fmt.Sprintf("Message #%d from producer at %s", counter, time.Now().Format(time.RFC3339))
		
		if err := p.PublishMessage(message); err != nil {
			log.Printf("Failed to publish message: %v", err)
			continue
		}
		
		log.Printf("Published: %s", message)
	}
}

func (p *Producer) Close() {
	if p.natsConn != nil {
		p.natsConn.Close()
	}
	if p.nacosClient != nil && p.serviceIP != "" {
		p.nacosClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          p.serviceIP,
			Port:        p.servicePort,
			ServiceName: ServiceName,
		})
	}
}

func main() {
	log.Println("Starting Producer Service...")

	producer, err := NewProducer(
		"nats://localhost:4222",
		"127.0.0.1",
		8848,
	)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	// 注册服务到 Nacos
	if err := producer.RegisterService("127.0.0.1", ServicePort); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	log.Println("Producer service is running...")
	producer.Start()
}
