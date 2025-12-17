package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/nats-io/nats.go"
)

const (
	ServiceName = "api-gateway"
	ServicePort = 8080
	NatsSubject = "microservice.events"
)

type Gateway struct {
	natsConn    *nats.Conn
	nacosClient naming_client.INamingClient
	serviceIP   string
	servicePort uint64
}

type MessageRequest struct {
	Content string `json:"content"`
}

type ServiceInfo struct {
	Name      string                 `json:"name"`
	Instances []model.Instance       `json:"instances"`
}

func NewGateway(natsURL, nacosAddr string, nacosPort uint64) (*Gateway, error) {
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

	return &Gateway{
		natsConn:    nc,
		nacosClient: namingClient,
	}, nil
}

func (g *Gateway) RegisterService(ip string, port uint64) error {
	success, err := g.nacosClient.RegisterInstance(vo.RegisterInstanceParam{
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
	g.serviceIP = ip
	g.servicePort = port

	log.Printf("Service %s registered successfully at %s:%d", ServiceName, ip, port)
	return nil
}

func (g *Gateway) PublishMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := g.natsConn.Publish(NatsSubject, []byte(req.Content)); err != nil {
		log.Printf("Failed to publish message: %v", err)
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}

	log.Printf("Published message via API: %s", req.Content)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Message published successfully",
	})
}

func (g *Gateway) ListServicesHandler(w http.ResponseWriter, r *http.Request) {
	serviceNames := []string{"producer-service", "consumer-service", "api-gateway"}
	services := []ServiceInfo{}

	for _, name := range serviceNames {
		instances, err := g.nacosClient.SelectInstances(vo.SelectInstancesParam{
			ServiceName: name,
			HealthyOnly: true,
		})

		if err != nil {
			log.Printf("Failed to get instances for %s: %v", name, err)
			continue
		}

		services = append(services, ServiceInfo{
			Name:      name,
			Instances: instances,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"services": services,
		"count":    len(services),
	})
}

func (g *Gateway) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

func (g *Gateway) Close() {
	if g.natsConn != nil {
		g.natsConn.Close()
	}
	if g.nacosClient != nil && g.serviceIP != "" {
		g.nacosClient.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          g.serviceIP,
			Port:        g.servicePort,
			ServiceName: ServiceName,
		})
	}
}

func main() {
	log.Println("Starting API Gateway...")

	gateway, err := NewGateway(
		"nats://localhost:4222",
		"127.0.0.1",
		8848,
	)
	if err != nil {
		log.Fatalf("Failed to create gateway: %v", err)
	}
	defer gateway.Close()

	// 注册服务到 Nacos
	if err := gateway.RegisterService("127.0.0.1", ServicePort); err != nil {
		log.Fatalf("Failed to register service: %v", err)
	}

	// 设置 HTTP 路由
	http.HandleFunc("/api/publish", gateway.PublishMessageHandler)
	http.HandleFunc("/api/services", gateway.ListServicesHandler)
	http.HandleFunc("/health", gateway.HealthHandler)

	log.Printf("API Gateway is running on http://127.0.0.1:%d", ServicePort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", ServicePort), nil))
}
