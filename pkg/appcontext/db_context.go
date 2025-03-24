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
	mu       sync.RWMutex // Мьютекс для безопасного доступа к instance
)

// SetInstance устанавливает глобальный экземпляр контекста базы данных
// Теперь позволяет переинициализировать соединения при необходимости
func SetInstance(postgresDB *connect.PostgresDB, redisClient *redis.Client) {
	mu.Lock()
	defer mu.Unlock()

	// Закрываем предыдущие соединения, если они существуют
	if instance != nil {
		instance.Close()
	}

	instance = &DbContext{
		PostgresDB:  postgresDB,
		RedisClient: redisClient,
	}
}

// GetInstance возвращает глобальный экземпляр контекста базы данных
func GetInstance() *DbContext {
	mu.RLock()
	defer mu.RUnlock()
	return instance
}

// Close закрывает все подключения к базам данных с безопасной синхронизацией
func (ctx *DbContext) Close() {
	// Используем локальный мьютекс вместо глобального
	// поскольку мы закрываем соединения конкретного экземпляра
	var wg sync.WaitGroup

	if ctx.PostgresDB != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx.PostgresDB.Close()
		}()
	}

	if ctx.RedisClient != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx.RedisClient.Close()
		}()
	}

	wg.Wait()
}

// CloseGlobalInstance безопасно закрывает глобальный экземпляр
func CloseGlobalInstance() {
	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		instance.Close()
		instance = nil
	}
}
