package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisCache struct {
	Client *redis.Client
}

func (r *RedisCache) getFromCache(ctx context.Context, searchID string) ([]interface{}, bool) {
	// Retrieve the search result from Redis
	val, err := r.Client.Get(ctx, searchID).Bytes()
	if err != nil {
		if err == redis.Nil {
			// Key does not exist in redisClient
			return nil, false
		}
		fmt.Println("Error retrieving result from redisClient:", err)
		return nil, false
	}

	var result []interface{}
	err = json.Unmarshal(val, &result)
	if err != nil {
		fmt.Println("Error deserializing result:", err)
		return nil, false
	}

	return result, true
}

func (r *RedisCache) saveToCache(ctx context.Context, searchID string, result interface{}, ttl time.Duration) {
	// Serialize the result to JSON
	data, err := json.Marshal(result)
	if err != nil {
		fmt.Println("Error serializing result:", err)
		return
	}

	// Set the search result in Redis with expiration time
	err = r.Client.Set(ctx, searchID, data, ttl).Err()
	if err != nil {
		fmt.Println("Error caching result:", err)
		return
	}
}
