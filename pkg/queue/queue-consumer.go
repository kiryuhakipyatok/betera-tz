package queue

import (
	"betera-tz/internal/config"
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	Client *kafka.Reader
	Config config.QueueConfig
}

func NewConsumer(cfg config.QueueConfig) *Consumer {
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{cfg.Broker},
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupId,
		CommitInterval: 0,
	})
	return &Consumer{
		Client: consumer,
		Config: cfg,
	}
}

func (c *Consumer) HandleMessages(handler MessageHandler) error {
	for {
		ctx := context.Background()
		msg, err := c.Client.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("faield to read message: %w", err)
		}

		message := Message{
			Key:   string(msg.Key),
			Value: msg.Value,
			Time:  msg.Time,
		}

		if err := handler(message); err != nil {
			log.Println("failed to handle message, will retry: %w", err)
			continue
		}

		if err := c.Client.CommitMessages(ctx, msg); err != nil {
			log.Println("failed to commit message: %w", err)
		}

	}
}

func (c *Consumer) MustClose() {
	if err := c.Client.Close(); err != nil {
		panic(fmt.Errorf("failed to close kafka consumer: %w", err))
	}
}
