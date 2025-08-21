package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type DbConf struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     string `envconfig:"DB_PORT" default:"5432"`
	User     string `envconfig:"DB_USER" default:"alan"`
	Password string `envconfig:"DB_PASSWORD" default:"2005"`
	DbName   string `envconfig:"DB_NAME" default:"orders_db"`
	Retries  int    `envconfig:"DB_CONNECTION_RETRIES" default:"3"`
}

type KafkaConf struct {
	Brokers   []string `envconfig:"KAFKA_BROKERS" default:"localhost:9092"`
	Topic     string   `envconfig:"KAFKA_TOPIC" default:"orders"`
	Partition int      `envconfig:"KAFKA_PARTITION" default:"0"`
	GroupID   string   `envconfig:"KAFKA_GROUP_ID" default:"orders-consumer"`
	Offset    int64    `envconfig:"KAFKA_OFFSET" default:"-1"`
}

type Config struct {
	DB        DbConf
	Kafka     KafkaConf
	CacheSize int    `envconfig:"CACHE_SIZE" default:"100"`
	HTTPPort  string `envconfig:"HTTP_PORT" default:"8000"`
	LogLevel  string `envconfig:"LOG_LEVEL" default:"debug"`
}

func Load() (Config, error) {
	var cfg Config
	_ = godotenv.Load()
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
