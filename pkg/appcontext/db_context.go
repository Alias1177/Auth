package appcontext

import (
	"Auth/internal/infrastructure/postgres/connect"
	"github.com/redis/go-redis/v9"
	"sync"
)

// DbContext хранит подключения к базам данных, доступные для всего приложения
type DbContext struct {
	PostgresDB  *connect.PostgresDB
	RedisClient *redis.Client
}

var (
	instance *DbContext
	once     sync.Once
	mu       sync.RWMutex
)

// SetInstance устанавливает глобальный экземпляр контекста базы данных
func SetInstance(postgresDB *connect.PostgresDB, redisClient *redis.Client) {
	once.Do(func() {
		instance = &DbContext{
			PostgresDB:  postgresDB,
			RedisClient: redisClient,
		}
	})
}

// GetInstance возвращает глобальный экземпляр контекста базы данных
func GetInstance() *DbContext {
	mu.RLock()
	defer mu.RUnlock()
	return instance
}

// Close закрывает все подключения к базам данных
func (ctx *DbContext) Close() {
	if ctx.PostgresDB != nil {
		ctx.PostgresDB.Close()
	}
	if ctx.RedisClient != nil {
		ctx.RedisClient.Close()
	}
}
