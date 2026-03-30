package config

import (
	"fmt"
	"strings"

	"github.com/wb-go/wbf/config"
)

type Config struct {
	HTTPAddr      string
	DatabaseDSN   string
	KafkaBrokers  []string
	KafkaTopic    string
	KafkaGroup    string
	StoragePath   string
	WatermarkText string
	WatermarkPath string
}

func Load() Config {
	c := config.New()
	err := c.LoadEnvFiles(".env")
	if err != nil {
		return Config{}
	}
	c.EnableEnv("APP")

	c.SetDefault("http.addr", "8080")
	c.SetDefault("db.dsn", "postgres://postgres:postgres@localhost:5432/image_processor?sslmode=disable")
	c.SetDefault("kafka.brokers", []string{"localhost:9091"})
	c.SetDefault("kafka.topic", "image-tasks")
	c.SetDefault("kafka.group", "")
	c.SetDefault("storage.path", "./storage")
	c.SetDefault("watermark.text", "watermark")
	c.SetDefault("watermark.path", "")

	return Config{
		HTTPAddr:      c.GetString("http.addr"),
		DatabaseDSN:   getDatabaseDSN(c),
		KafkaBrokers:  strings.Split(c.GetString("kafka.brokers"), ","),
		KafkaTopic:    c.GetString("kafka.topic"),
		KafkaGroup:    c.GetString("kafka.group"),
		StoragePath:   c.GetString("storage.path"),
		WatermarkText: c.GetString("watermark.text"),
		WatermarkPath: c.GetString("watermark.path"),
	}
}

func getDatabaseDSN(c *config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.GetString("postgres.user"),
		c.GetString("postgres.pass"),
		c.GetString("postgres.host"),
		c.GetString("postgres.port"),
		c.GetString("postgres.db"),
		c.GetString("postgres.ssl.mode"))
}
