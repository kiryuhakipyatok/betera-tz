package queue

import (
	"betera-tz/internal/config"
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	Client *kafka.Writer
	Config config.QueueConfig
}

func NewProducer(cfg config.QueueConfig) *Producer {
	p := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{cfg.Broker},
		Topic:        cfg.Topic,
		RequiredAcks: 1,
		Balancer:     &kafka.RoundRobin{},
	})
	return &Producer{
		Client: p,
		Config: cfg,
	}
}

func (p *Producer) SendMessage(message Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.Config.Timeout)
	defer cancel()

	kafkaMsg := kafka.Message{
		Key:   []byte(message.Key),
		Value: message.Value,
		Time:  message.Time,
	}

	if err := p.Client.WriteMessages(ctx, kafkaMsg); err != nil {
		return err
	}
	return nil
}

func (p *Producer) MustClose() {
	if err := p.Client.Close(); err != nil {
		panic(fmt.Errorf("failed to close kafka producer: %w", err))
	}
}
