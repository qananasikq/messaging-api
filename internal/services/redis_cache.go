package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"messaging-api/internal/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisCacheConfig struct {
	LastMessagesLimit int
	LastMessagesTTL   time.Duration
	UnreadTTL         time.Duration
}

type RedisCache struct {
	rdb *redis.Client
	cfg RedisCacheConfig
}

func NewRedisCache(rdb *redis.Client, cfg RedisCacheConfig) *RedisCache {
	if cfg.LastMessagesLimit <= 0 {
		cfg.LastMessagesLimit = 100
	}
	if cfg.LastMessagesTTL <= 0 {
		cfg.LastMessagesTTL = 15 * time.Minute
	}
	if cfg.UnreadTTL <= 0 {
		cfg.UnreadTTL = 20 * time.Minute
	}
	return &RedisCache{rdb: rdb, cfg: cfg}
}

func (c *RedisCache) lastMessagesKey(dialogID uuid.UUID) string {
	return fmt.Sprintf("dialog:%s:last_messages", dialogID.String())
}

func (c *RedisCache) unreadKey(userID, dialogID uuid.UUID) string {
	return fmt.Sprintf("unread:%s:%s", userID.String(), dialogID.String())
}

func (c *RedisCache) PushLastMessage(ctx context.Context, dialogID uuid.UUID, msg models.Message) error {
	key := c.lastMessagesKey(dialogID)
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	pipe := c.rdb.Pipeline()
	pipe.LPush(ctx, key, b)
	pipe.LTrim(ctx, key, 0, int64(c.cfg.LastMessagesLimit-1))
	pipe.Expire(ctx, key, c.cfg.LastMessagesTTL)
	_, err = pipe.Exec(ctx)
	return err
}

func (c *RedisCache) GetLastMessages(ctx context.Context, dialogID uuid.UUID, limit int) ([]models.Message, bool, error) {
	if limit <= 0 || limit > c.cfg.LastMessagesLimit {
		limit = c.cfg.LastMessagesLimit
	}
	key := c.lastMessagesKey(dialogID)

	items, err := c.rdb.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if len(items) == 0 {
		return nil, false, nil
	}

	out := make([]models.Message, 0, len(items))
	for _, s := range items {
		var m models.Message
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			return nil, false, nil
		}
		out = append(out, m)
	}
	return out, true, nil
}

func (c *RedisCache) IncrUnreadForUsers(ctx context.Context, dialogID uuid.UUID, userIDs []uuid.UUID) error {
	pipe := c.rdb.Pipeline()
	for _, uid := range userIDs {
		k := c.unreadKey(uid, dialogID)
		pipe.Incr(ctx, k)
		pipe.Expire(ctx, k, c.cfg.UnreadTTL)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisCache) GetUnread(ctx context.Context, userID, dialogID uuid.UUID) (int64, bool, error) {
	k := c.unreadKey(userID, dialogID)
	n, err := c.rdb.Get(ctx, k).Int64()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return n, true, nil
}

func (c *RedisCache) ResetUnread(ctx context.Context, userID, dialogID uuid.UUID) error {
	k := c.unreadKey(userID, dialogID)
	pipe := c.rdb.Pipeline()
	pipe.Set(ctx, k, 0, c.cfg.UnreadTTL)
	_, err := pipe.Exec(ctx)
	return err
}
