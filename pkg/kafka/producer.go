package kafka

import (
	"Auth/pkg/logger"
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"time"
)

// Producer представляет собой клиент для отправки сообщений в Kafka
type Producer struct {
	writer *kafka.Writer
	logger *logger.Logger
}

// NewProducer создает новый экземпляр Producer
func NewProducer(brokerAddress, topic string, log *logger.Logger) *Producer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{brokerAddress},
		Topic:        topic, // Используем "emails" вместо "user_registrations"
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		Async:        false,
	})

	log.Infow("Kafka producer initialized", "broker", brokerAddress, "topic", topic)

	return &Producer{
		writer: writer,
		logger: log,
	}
}

// SendEmailRegistration отправляет email адрес пользователя в Kafka
// Теперь отправляем только email как строку
func (p *Producer) SendEmailRegistration(ctx context.Context, email, username string) error {
	// Отправляем только email адрес как строку
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Value: []byte(email),
		Key:   []byte(email), // Используем email как ключ
	})

	if err != nil {
		p.logger.Errorw("Failed to send message to Kafka", "error", err)
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	p.logger.Infow("Email sent to Kafka", "email", email)
	return nil
}

// Close закрывает соединение с Kafka
func (p *Producer) Close() error {
	if p.writer != nil {
		err := p.writer.Close()
		if err != nil {
			p.logger.Errorw("Failed to close Kafka producer", "error", err)
			return fmt.Errorf("failed to close Kafka producer: %w", err)
		}
	}
	return nil
}
