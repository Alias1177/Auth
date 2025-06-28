package kafka

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Alias1177/Auth/pkg/logger"
)

// MockKafkaWriter реализует интерфейс Writer
type MockKafkaWriter struct {
	mock.Mock
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockLogger реализует logger.Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Infow(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Errorw(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) WithFields(fields map[string]interface{}) logger.Logger {
	args := m.Called(fields)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewProducer(t *testing.T) {
	tests := []struct {
		name    string
		broker  string
		topic   string
		wantErr bool
	}{
		{
			name:    "Success",
			broker:  "localhost:9092",
			topic:   "emails",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, err := logger.NewSimpleLogger("info")
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			got := NewProducer(tt.broker, tt.topic, log)

			if !tt.wantErr {
				assert.NotNil(t, got)
				assert.Equal(t, log, got.logger)

				writer := reflect.ValueOf(got.writer).Elem()
				assert.Equal(t, tt.topic, writer.FieldByName("Topic").String())
				assert.Equal(t, 1, int(writer.FieldByName("BatchSize").Int()))
				assert.Equal(t, 10*time.Millisecond, writer.FieldByName("BatchTimeout").Interface().(time.Duration))
			}
		})
	}
}

func TestProducer_SendEmailRegistration(t *testing.T) {
	type args struct {
		ctx      context.Context
		email    string
		username string
	}

	tests := []struct {
		name    string
		args    args
		mock    func(*MockKafkaWriter, *MockLogger)
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				ctx:      context.Background(),
				email:    "test@example.com",
				username: "user",
			},
			mock: func(w *MockKafkaWriter, l *MockLogger) {
				w.On("WriteMessages", context.Background(), []kafka.Message{
					{Value: []byte("test@example.com"), Key: []byte("test@example.com")},
				}).Return(nil)
				l.On("Infow", "Email sent to Kafka", []interface{}{"email", "test@example.com"}).Return()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := &MockKafkaWriter{}
			mockLogger := &MockLogger{}

			// Создаем Producer с подмененными зависимостями через reflection
			p := &Producer{
				writer: &kafka.Writer{},
				logger: &logger.Logger{},
			}

			reflect.ValueOf(p).Elem().FieldByName("writer").Set(reflect.ValueOf(mockWriter))
			reflect.ValueOf(p).Elem().FieldByName("logger").Set(reflect.ValueOf(mockLogger))

			tt.mock(mockWriter, mockLogger)

			err := p.SendEmailRegistration(tt.args.ctx, tt.args.email, tt.args.username)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockWriter.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestProducer_Close(t *testing.T) {
	tests := []struct {
		name    string
		mock    func(*MockKafkaWriter, *MockLogger)
		wantErr bool
	}{
		{
			name: "Success",
			mock: func(w *MockKafkaWriter, l *MockLogger) {
				w.On("Close").Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := &MockKafkaWriter{}
			mockLogger := &MockLogger{}

			// Создаем Producer с подмененными зависимостями через reflection
			p := &Producer{
				writer: &kafka.Writer{},
				logger: &logger.Logger{},
			}

			reflect.ValueOf(p).Elem().FieldByName("writer").Set(reflect.ValueOf(mockWriter))
			reflect.ValueOf(p).Elem().FieldByName("logger").Set(reflect.ValueOf(mockLogger))

			tt.mock(mockWriter, mockLogger)

			err := p.Close()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockWriter.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
