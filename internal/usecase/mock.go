package usecase

import (
	"Auth/internal/entity"
	"context"

	"github.com/stretchr/testify/mock"
)

// --- MockUserRepository ---
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	args := m.Called(ctx, id)
	user, ok := args.Get(0).(*entity.User)
	if !ok && args.Get(0) != nil {
		return nil, args.Error(1)
	}
	return user, args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	user, ok := args.Get(0).(*entity.User)
	if !ok && args.Get(0) != nil {
		return nil, args.Error(1)
	}
	return user, args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// --- MockTokenManager ---
type MockTokenManager struct {
	mock.Mock
}

func (m *MockTokenManager) GenerateAccessToken(userClaims entity.UserClaims) (string, error) {
	args := m.Called(userClaims)
	return args.String(0), args.Error(1)
}

func (m *MockTokenManager) ValidateAccessToken(token string) (*entity.UserClaims, error) {
	args := m.Called(token)
	claims, ok := args.Get(0).(*entity.UserClaims)
	if !ok && args.Get(0) != nil {
		return nil, args.Error(1)
	}
	return claims, args.Error(1)
}

// --- MockLogger ---
var _ Logger = (*MockLogger)(nil)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Infow(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Errorw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Warnw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Debugw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Fatalw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Close() error {
	args := m.Called()
	return args.Error(0)
}

// --- MockUserCache ---
type MockUserCache struct {
	mock.Mock
}

func (m *MockUserCache) GetUser(ctx context.Context, id int) (*entity.User, error) {
	args := m.Called(ctx, id)
	user, ok := args.Get(0).(*entity.User)
	if !ok && args.Get(0) != nil {
		return nil, args.Error(1)
	}
	return user, args.Error(1)
}

func (m *MockUserCache) SaveUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// --- MockKafkaProducer ---
type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockKafkaProducer) Send(ctx context.Context, topic string, message any) error {
	args := m.Called(ctx, topic, message)
	return args.Error(0)
}
