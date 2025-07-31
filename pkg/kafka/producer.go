package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Alias1177/Auth/pkg/logger"
	"github.com/segmentio/kafka-go"
)

// PasswordResetRequest структура для запроса восстановления пароля
type PasswordResetRequest struct {
	Email  string `json:"email"`
	UserID string `json:"user_id,omitempty"`
}

// RegistrationRequest структура для запроса регистрации
type RegistrationRequest struct {
	Email    string `json:"email"`
	Username string `json:"username,omitempty"`
}

// Producer представляет собой клиент для отправки сообщений в Kafka
type Producer struct {
	writer *kafka.Writer
	logger *logger.Logger
}

// NewProducer создает новый экземпляр Producer
func NewProducer(brokerAddress, topic string, log *logger.Logger) *Producer {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{brokerAddress},
		Topic:        topic,
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

// SendPasswordResetRequest отправляет запрос на восстановление пароля
func (p *Producer) SendPasswordResetRequest(ctx context.Context, email string) error {
	request := PasswordResetRequest{
		Email: email,
	}

	data, err := json.Marshal(request)
	if err != nil {
		p.logger.Errorw("Failed to marshal password reset request", "error", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Value: data,
		Key:   []byte(email),
	})

	if err != nil {
		p.logger.Errorw("Failed to send password reset request", "error", err)
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	p.logger.Infow("Password reset request sent to Kafka", "email", email)
	return nil
}

// SendRegistrationRequest отправляет запрос на регистрацию
func (p *Producer) SendRegistrationRequest(ctx context.Context, email, username string) error {
	request := RegistrationRequest{
		Email:    email,
		Username: username,
	}

	data, err := json.Marshal(request)
	if err != nil {
		p.logger.Errorw("Failed to marshal registration request", "error", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Value: data,
		Key:   []byte(email),
	})

	if err != nil {
		p.logger.Errorw("Failed to send registration request", "error", err)
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	p.logger.Infow("Registration request sent to Kafka", "email", email)
	return nil
}

// SendEmailRegistration отправляет email адрес пользователя в Kafka (для обратной совместимости)
// Теперь отправляем только email как строку
func (p *Producer) SendEmailRegistration(ctx context.Context, email, username string) error {
	// Используем новую структурированную функцию
	return p.SendRegistrationRequest(ctx, email, username)
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
