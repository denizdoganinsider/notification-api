package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"notification-api/internal/domain"

	"github.com/redis/go-redis/v9"
)

var errNoRedis = errors.New("redis client not configured")

type CacheService struct {
	redisClient *redis.Client
	ttl         time.Duration
}

func NewCacheService(redisClient *redis.Client) *CacheService {
	return &CacheService{
		redisClient: redisClient,
		ttl:         5 * time.Minute,
	}
}

func (s *CacheService) GetNotifications(userID int64, page int, perPage int) ([]domain.Notification, error) {
	if s.redisClient == nil {
		return nil, errNoRedis
	}

	ctx := context.Background()
	key := fmt.Sprintf("notifications:user:%d:page:%d:per_page:%d", userID, page, perPage)

	data, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var notifications []domain.Notification
	err = json.Unmarshal([]byte(data), &notifications)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (s *CacheService) SetNotifications(userID int64, page int, perPage int, notifications []domain.Notification) error {
	if s.redisClient == nil {
		return errNoRedis
	}

	ctx := context.Background()
	key := fmt.Sprintf("notifications:user:%d:page:%d:per_page:%d", userID, page, perPage)

	data, err := json.Marshal(notifications)
	if err != nil {
		return err
	}

	return s.redisClient.Set(ctx, key, data, s.ttl).Err()
}

func (s *CacheService) InvalidateUserNotifications(userID int64) error {
	if s.redisClient == nil {
		return errNoRedis
	}

	ctx := context.Background()
	pattern := fmt.Sprintf("notifications:user:%d:*", userID)

	iter := s.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		s.redisClient.Del(ctx, iter.Val())
	}

	return iter.Err()
}
