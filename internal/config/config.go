package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
// Use env tags for env var names and defaults
// Use envDefault for default values, required for required fields
// Use envSeparator for slices

type Config struct {
	App      AppConfig      `envPrefix:"APP_"`
	Database DatabaseConfig `envPrefix:"DATABASE_"`
	Log      LogConfig      `envPrefix:"LOG_"`
	Auth     AuthConfig     `envPrefix:"AUTH_"`
	Redis    RedisConfig    `envPrefix:"REDIS_"`
	Kafka    KafkaConfig    `envPrefix:"KAFKA_"`
}

type AppConfig struct {
	Name    string `env:"NAME" envDefault:"microservice-boilerplate"`
	Version string `env:"VERSION" envDefault:"1.0.0"`
	Port    int    `env:"PORT" envDefault:"8080"`
	Env     string `env:"ENV" envDefault:"development"`
}

type DatabaseConfig struct {
	Host         string `env:"HOST" envDefault:"host.docker.internal"`
	Port         int    `env:"PORT" envDefault:"5432"`
	Name         string `env:"NAME" envDefault:"microservice_db"`
	User         string `env:"USER" envDefault:"postgres"`
	Password     string `env:"PASSWORD" envDefault:"password"`
	SSLMode      string `env:"SSL_MODE" envDefault:"disable"`
	MaxOpenConns int    `env:"MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns int    `env:"MAX_IDLE_CONNS" envDefault:"5"`
}

type LogConfig struct {
	Level  string `env:"LEVEL" envDefault:"info"`
	Format string `env:"FORMAT" envDefault:"json"`
}

type AuthConfig struct {
	JWTSecret     string        `env:"JWT_SECRET" envDefault:"your-super-secret-jwt-key-change-in-production"`
	RefreshSecret string        `env:"REFRESH_SECRET" envDefault:"your-super-secret-refresh-key-change-in-production"`
	AccessTTL     time.Duration `env:"ACCESS_TTL" envDefault:"15m"`
	RefreshTTL    time.Duration `env:"REFRESH_TTL" envDefault:"168h"`
}

type RedisConfig struct {
	Host     string `env:"HOST" envDefault:"host.docker.internal"`
	Port     int    `env:"PORT" envDefault:"6379"`
	Password string `env:"PASSWORD" envDefault:""`
	DB       int    `env:"DB" envDefault:"0"`
}

type KafkaConfig struct {
	Brokers      []string `env:"BROKERS" envDefault:"host.docker.internal:9092" envSeparator:","`
	TopicPrefix  string   `env:"TOPIC_PREFIX" envDefault:"microservice"`
	GroupID      string   `env:"GROUP_ID" envDefault:"microservice-group"`
	AutoOffset   string   `env:"AUTO_OFFSET" envDefault:"earliest"`
	MaxRetries   int      `env:"MAX_RETRIES" envDefault:"3"`
	RetryBackoff int      `env:"RETRY_BACKOFF" envDefault:"1000"`
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	_ = godotenv.Load()
	var config Config
	if err := env.Parse(&config); err != nil {
		return nil, fmt.Errorf("failed to parse env: %w", err)
	}
	fmt.Println(config)
	return &config, nil
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

// GetRedisAddr returns the Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
