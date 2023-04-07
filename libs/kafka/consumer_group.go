package kafka

import (
	"context"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
)

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready    chan bool
	handlers map[string]func([]byte)
}

type ConsumerGroup struct {
	consumer Consumer
	brokers  []string
	topics   []string
	name     string
	strategy string
}

// NewConsumerGroup - constructor
func NewConsumerGroup(handlers map[string]func([]byte), brokers, topics []string, name, strategy string) *ConsumerGroup {
	return &ConsumerGroup{
		consumer: Consumer{
			ready:    make(chan bool),
			handlers: handlers,
		},
		brokers:  brokers,
		topics:   topics,
		name:     name,
		strategy: strategy,
	}
}

func (cg *ConsumerGroup) Run(ctx context.Context) error {
	config := sarama.NewConfig()
	config.Version = sarama.MaxVersion
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	switch cg.strategy {
	case "sticky":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategySticky}
	case "roundrobin":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRoundRobin}
	default:
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRange}
	}
	client, err := sarama.NewConsumerGroup(cg.brokers, cg.name, config)
	if err != nil {
		return errors.Wrap(err, "Error creating consumer group client")
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := client.Consume(ctx, cg.topics, &cg.consumer); err != nil {
				return
			}
			if ctx.Err() != nil {
				return
			}
			cg.consumer.ready = make(chan bool)
		}
	}()

	<-cg.consumer.ready
	<-ctx.Done()
	wg.Wait()
	if err = client.Close(); err != nil {
		return errors.Wrap(err, "Error closing client: %v")
	}
	return nil
}

func (c *Consumer) Ready() <-chan bool {
	return c.ready
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			handler, ok := c.handlers[message.Topic]
			if !ok {
				return errors.New("no handler for topic")
			}
			handler(message.Value)
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}
