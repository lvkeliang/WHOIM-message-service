package main

import (
	"fmt"
	"github.com/lvkeliang/WHOIM-message-service/dao"
	"github.com/lvkeliang/WHOIM-message-service/service"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	dao.InitRPC()

	dao.SetRocketMQLogLevel()

	// 初始化 RocketMQ 生产者
	err := dao.InitMQProducer()
	if err != nil {
		log.Fatalf("Failed to initialize RocketMQ producer: %v", err)
	}

	// 启动消息队列监听
	err = service.StartListening()
	if err != nil {
		log.Fatalf("Failed to start message queue listener: %v", err)
	}

	// 捕捉信号并优雅关闭服务
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	sig := <-sigCh
	fmt.Printf("Received signal: %v, shutting down...\n", sig)

	dao.ShutdownProducer()
	service.StopListeningForService()
}
