package service

import (
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/lvkeliang/WHOIM-message-service/dao"
	"log"
)

var globalConsumer rocketmq.PushConsumer

// StartListening 启动指定 serviceID 的消息队列监听
func StartListening() error {

	// 调用 InitMQConsumer 来启动消息监听，传入 messageHandler 作为处理函数
	consumer, err := dao.InitMQConsumer(messageHandler)
	if err != nil {
		log.Printf("Failed to listen to consumer queue for service: %v", err)
		return err
	}

	globalConsumer = consumer

	log.Printf("Started listening to message queue for service successfully")
	return nil
}

// StopListeningForService 停止指定 serviceID 的消息队列监听
func StopListeningForService() {

	// 调用 ShutdownConsumer 来关闭消费者
	dao.ShutdownConsumer(globalConsumer)

}
