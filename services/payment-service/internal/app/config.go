package app

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                  string
	KafkaBrokers          string
	KafkaAutoOffsetReset  string
	PaymentProcessTimeout time.Duration
	KafkaPublishTimeout   time.Duration
	RedisAddr             string
	RedisDB               int
	OrderTotalTTL         time.Duration
}

func LoadConfigFromEnv() *Config {
	procTimeout := 5 * time.Second
	if v := os.Getenv("PAYMENT_PROCESS_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			procTimeout = d
		}
	}
	pubTimeout := 3 * time.Second
	if v := os.Getenv("KAFKA_PUBLISH_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			pubTimeout = d
		}
	}
	redisDB := 0
	if v := os.Getenv("REDIS_DB"); v != "" {
		if d, err := strconv.Atoi(v); err == nil {
			redisDB = d
		}
	}
	orderTTL := 30 * time.Minute
	if v := os.Getenv("ORDER_TOTAL_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			orderTTL = d
		}
	}
	return &Config{
		Port:                  getEnv("PORT", "50054"),
		KafkaBrokers:          getEnv("KAFKA_BROKERS", "kafka:9092"),
		KafkaAutoOffsetReset:  getEnv("KAFKA_AUTO_OFFSET_RESET", "earliest"),
		PaymentProcessTimeout: procTimeout,
		KafkaPublishTimeout:   pubTimeout,
		RedisAddr:             getEnv("REDIS_ADDR", "redis:6379"),
		RedisDB:               redisDB,
		OrderTotalTTL:         orderTTL,
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
