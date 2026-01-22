package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var ctx = context.Background()

// InitRedis conecta a Redis para caching
func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	// Verificar conexión
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("error conectando a Redis: %w", err)
	}

	log.Println("✅ Conectado a Redis")
	return nil
}

// CacheSet guarda un valor en cache con TTL
func CacheSet(key string, value interface{}, ttl time.Duration) error {
	json, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return RedisClient.Set(ctx, key, json, ttl).Err()
}

// CacheGet obtiene un valor del cache
func CacheGet(key string, dest interface{}) error {
	val, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// CacheDelete elimina una clave del cache
func CacheDelete(key string) error {
	return RedisClient.Del(ctx, key).Err()
}

// CacheExists verifica si una clave existe
func CacheExists(key string) bool {
	val, err := RedisClient.Exists(ctx, key).Result()
	return err == nil && val > 0
}
