package service

import (
	"context"
	"encoding/json"
	"github.com/lvkeliang/WHOIM-message-service/config"
	"github.com/lvkeliang/WHOIM-message-service/dao"
	"github.com/lvkeliang/WHOIM-message-service/protocol"
	"log"
)

// SendMessageToTopic 向指定用户的所有设备发送消息
func SendMessageToTopic(userID string, message *protocol.MessageProtocol) error {
	devices, err := dao.UserClient.GetUserDevices(context.Background(), userID)
	if devices == nil {
		log.Printf("User %s has no connected devices", userID)
		return nil
	}

	// 将 MessageProtocol 结构体转换为 JSON
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return err
	}

	cfg := config.LoadConfig()
	// 用一个 map 来记录已发送的 ServerAddress，防止重复发送
	sentServers := make(map[string]bool)

	// 遍历所有设备状态
	for _, status := range devices {
		// 如果该 ServerAddress 尚未发送
		if _, exists := sentServers[status.ServerAddress]; !exists {
			// 发送消息到 RocketMQ 单发消息队列
			err = dao.SendMessage(messageBytes, cfg.RocketMQServerConsumerGroupName+"_"+status.ServerAddress)
			if err != nil {
				log.Printf("Failed to send message to %s: %v", status.ServerAddress, err)
			} else {
				log.Printf("Message sent to RocketMQ successfully for ServerAddress: %s", status.ServerAddress)
			}
			// 标记这个 ServerAddress 已发送
			sentServers[status.ServerAddress] = true
		}
	}

	return nil
}
