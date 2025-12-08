package repository

import (
	"sync"

	"github.com/perhydrol/insurance-agent-backend/internal/infrastructure/cache"
	"github.com/perhydrol/insurance-agent-backend/pkg/domain"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Repo struct {
	OrderRepo   domain.OrderRepository
	UserRepo    domain.UserRepository
	ProductRepo domain.ProductRepository
	ChatRepo    domain.ChatRepository
}

var Repository Repo
var once sync.Once

func InitRepository(db *gorm.DB, redis *redis.Client) {
	once.Do(func() {
		Repository = Repo{
			OrderRepo:   NewOrderRepo(db, cache.NewOrderCache(redis), cache.NewRedisLocker(redis)),
			UserRepo:    NewUerRepository(db, cache.NewUserCache(redis)),
			ProductRepo: NewProductRepository(db, cache.NewProductCache(redis)),
			ChatRepo:    NewChatRepository(db, cache.NewChatMessageCache(redis)),
		}
	})
}
