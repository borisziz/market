package main

import (
	"context"
	"log"
	"route256/libs/kafka"
	"route256/notifications/internal/config"
	"route256/notifications/internal/domain"
)

func main() {
	err := config.Init()
	if err != nil {
		log.Fatal("config init", err)
	}
	d := domain.New()
	handlers := make(map[string]func([]byte), len(config.ConfigData.Kafka.Topics))
	for _, topic := range config.ConfigData.Kafka.Topics {
		handlers[topic] = d.ReceiveOrder
	}
	cg := kafka.NewConsumerGroup(handlers, config.ConfigData.Kafka.Brokers, config.ConfigData.Kafka.Topics, config.ConfigData.Kafka.GroupName, config.ConfigData.Kafka.Strategy)
	cg.Run(context.Background())
}
