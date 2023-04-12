package main

import (
	"context"
	"go.uber.org/zap"
	"route256/libs/kafka"
	"route256/libs/logger"
	"route256/notifications/internal/config"
	"route256/notifications/internal/domain"
)

func main() {
	logger.Init(true)
	err := config.Init()
	if err != nil {
		logger.Fatal("config init", zap.Error(err))
	}
	d := domain.New()
	handlers := make(map[string]func([]byte), len(config.ConfigData.Kafka.Topics))
	for _, topic := range config.ConfigData.Kafka.Topics {
		handlers[topic] = d.ReceiveOrder
	}
	cg := kafka.NewConsumerGroup(handlers, config.ConfigData.Kafka.Brokers, config.ConfigData.Kafka.Topics, config.ConfigData.Kafka.GroupName, config.ConfigData.Kafka.Strategy)
	logger.Info("waiting notifications")
	err = cg.Run(context.Background())
	if err != nil {
		logger.Fatal("wait notifications", zap.Error(err))
	}
}
