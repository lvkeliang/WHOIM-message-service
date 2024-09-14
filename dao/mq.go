package dao

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/lvkeliang/WHOIM-message-service/config"
	"log"
)

var (
	// Global variables for producer and consumer
	MQProducer rocketmq.Producer
)

func SetRocketMQLogLevel() {
	// 将日志级别设置为 WARN 级别，减少日志输出
	rlog.SetLogLevel("warn")
}

func CreateTopic(topicName string) {
	cfg := config.LoadConfig()

	endPoint := []string{cfg.RocketMQNameSrv}
	// 创建主题
	testAdmin, err := admin.NewAdmin(admin.WithResolver(primitive.NewPassthroughResolver(endPoint)))
	if err != nil {
		fmt.Printf("connection error: %s\n", err.Error())
	}

	brokerAddress := cfg.RocketMQBrokerAddress

	err = testAdmin.CreateTopic(context.Background(), admin.WithTopicCreate(topicName), admin.WithBrokerAddrCreate(brokerAddress))
	if err != nil {
		fmt.Printf("createTopic error: %s\n", err.Error())
	}
}

// DeleteTopic 不建议调用,应当使用进程监测服务在检测到进程关闭时来清理并关闭topic
func DeleteTopic(topicName string) {
	cfg := config.LoadConfig()

	endPoint := []string{cfg.RocketMQNameSrv}

	// 创建 Admin 实例
	testAdmin, err := admin.NewAdmin(admin.WithResolver(primitive.NewPassthroughResolver(endPoint)))
	if err != nil {
		fmt.Printf("connection error: %s\n", err.Error())
		return
	}

	// 获取 Broker 地址
	brokerAddress := cfg.RocketMQBrokerAddress

	// 调用删除 topic 的接口
	err = testAdmin.DeleteTopic(context.Background(), admin.WithTopicDelete(topicName), admin.WithBrokerAddrDelete(brokerAddress))
	if err != nil {
		fmt.Printf("deleteTopic error: %s\n", err.Error())
		return
	}

	fmt.Printf("Topic %s deleted successfully\n", topicName)
}

func InitMQProducer() error {
	cfg := config.LoadConfig()

	var err error
	MQProducer, err = rocketmq.NewProducer(
		producer.WithGroupName(cfg.RocketMQSingleMessageProducerGroupName),
		producer.WithNameServer([]string{cfg.RocketMQNameSrv}),
	)

	if err != nil {
		log.Fatalf("Create RocketMQ producer error: %s", err)
		return err
	}

	err = MQProducer.Start()
	if err != nil {
		log.Fatalf("Start RocketMQ producer error: %s", err)
		return err
	}

	log.Println("RocketMQ producer started successfully")
	return nil
}

// ShutdownProducer shuts down the producer
func ShutdownProducer() {
	if MQProducer != nil {
		err := MQProducer.Shutdown()
		if err != nil {
			log.Printf("Shutdown producer error: %s", err.Error())
		} else {
			fmt.Println("RocketMQ Producer shutdown successfully")
		}
	}
}

type MessageExt = func(context.Context, ...*primitive.MessageExt) (consumer.ConsumeResult, error)

func InitMQConsumer(f MessageExt) (rocketmq.PushConsumer, error) {
	cfg := config.LoadConfig()

	endPoint := []string{cfg.RocketMQNameSrv}

	consumerGroup := cfg.RocketMQSingleMessageConsumerGroupName

	// 创建重试Topic
	err := CreateRetryTopicForGroup(consumerGroup)
	if err != nil {
		log.Fatalf("Failed to create retry topic: %s", err)
		return nil, err
	}

	// 创建一个consumer实例
	MQConsumer, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer(endPoint),
		consumer.WithConsumerModel(consumer.Clustering),
		consumer.WithGroupName(consumerGroup),
	)
	if err != nil {
		log.Fatalf("Create consumer error: %s", err.Error())
		return nil, err
	}

	// 订阅topic
	err = MQConsumer.Subscribe(cfg.RocketMQSingleMessageTopic, consumer.MessageSelector{}, f)
	if err != nil {
		log.Fatalf("Subscribe topic error: %s", err.Error())
		return nil, err
	}

	// 启动consumer
	err = MQConsumer.Start()
	if err != nil {
		log.Fatalf("Start consumer error: %s", err.Error())
		return nil, err
	}

	fmt.Println("RocketMQ Consumer started successfully for topic:", cfg.RocketMQSingleMessageTopic)
	return MQConsumer, nil
}

// ShutdownConsumer shuts down the specified consumer
func ShutdownConsumer(consumer rocketmq.PushConsumer) {
	if consumer != nil {
		err := consumer.Shutdown()
		if err != nil {
			log.Printf("Shutdown consumer error: %s", err.Error())
		} else {
			fmt.Println("RocketMQ Consumer shutdown successfully")
		}
	}
}

func SendMessage(message []byte, topic string) error {

	// 发送消息
	result, err := MQProducer.SendSync(context.Background(), &primitive.Message{
		Topic: topic,
		Body:  message,
	})

	if err != nil {
		log.Printf("Send message error: %v", err)
		return err
	}

	log.Printf("Send message success: result=%v, messageID=%s", result.Status, result.MsgID)
	return nil
}

func CreateRetryTopicForGroup(consumerGroup string) error {
	cfg := config.LoadConfig()
	endPoint := []string{cfg.RocketMQNameSrv}

	adminClient, err := admin.NewAdmin(admin.WithResolver(primitive.NewPassthroughResolver(endPoint)))
	if err != nil {
		log.Fatalf("Create Admin error: %s", err.Error())
		return err
	}

	// 创建重试Topic
	retryTopic := fmt.Sprintf("%%RETRY%%%s", consumerGroup)
	brokerAddress := cfg.RocketMQBrokerAddress

	err = adminClient.CreateTopic(context.Background(), admin.WithTopicCreate(retryTopic), admin.WithBrokerAddrCreate(brokerAddress))
	if err != nil {
		log.Fatalf("Create retry topic error: %s", err.Error())
		return err
	}

	log.Printf("Retry topic for consumer group %s created successfully", consumerGroup)
	return nil
}
