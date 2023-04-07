package sender

import (
	"fmt"
	"log"
	"route256/loms/internal/api/loms/v1"
	"route256/loms/internal/domain"
	desc "route256/loms/pkg/loms/v1"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
)

var _ domain.NotificationsSender = (*orderSender)(nil)

type orderSender struct {
	producer sarama.SyncProducer
	topic    string
}

func NewOrderSender(producer sarama.SyncProducer, topic string) *orderSender {
	return &orderSender{
		producer: producer,
		topic:    topic,
	}
}

func (s *orderSender) SendOrder(order *domain.Order) error {
	orderpb := &desc.Order{
		Id:     order.ID,
		Status: loms.StatusToStatusCode(order.Status),
		User:   order.User,
	}
	items := make([]*desc.Item, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, &desc.Item{
			Sku:   item.Sku,
			Count: uint32(item.Count),
		})
	}
	orderpb.Items = items
	bytes, err := protojson.Marshal(orderpb)
	if err != nil {
		return errors.Wrap(err, "marshal order")
	}

	msg := &sarama.ProducerMessage{
		Topic:     s.topic,
		Partition: -1,
		Value:     sarama.ByteEncoder(bytes),
		Key:       sarama.StringEncoder(fmt.Sprint(order.ID)),
		Timestamp: time.Now(),
	}

	partition, offset, err := s.producer.SendMessage(msg)
	if err != nil {
		return errors.Wrap(err, "send message")
	}

	log.Printf("order id: %d, partition: %d, offset: %d", order.ID, partition, offset)
	return nil
}
