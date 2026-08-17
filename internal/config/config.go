package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	GRPC     GRPCConfig     `mapstructure:"grpc"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Mongo    MongoConfig    `mapstructure:"mongo"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Outbox   OutboxConfig   `mapstructure:"outbox"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	External ExternalConfig `mapstructure:"external"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Env     string `mapstructure:"env"`
	Secret  string `mapstructure:"secret"`
	Version string `mapstructure:"version"`
}

type HTTPConfig struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type GRPCConfig struct {
	Port            string        `mapstructure:"port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type PostgresConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode,
	)
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MongoConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

type KafkaConfig struct {
	Brokers           []string `mapstructure:"brokers"`
	GroupID           string   `mapstructure:"group_id"`
	Topic             string   `mapstructure:"topic"`
	DLQTopic          string   `mapstructure:"dlq_topic"`
	MaxHandlerRetries int      `mapstructure:"max_handler_retries"`
}

type OutboxConfig struct {
	PollInterval         time.Duration `mapstructure:"poll_interval"`
	BatchSize            int           `mapstructure:"batch_size"`
	MaxRetries           int           `mapstructure:"max_retries"`
	ProcessingStaleAfter time.Duration `mapstructure:"processing_stale_after"`
}

type ObservabilityConfig struct {
	LogLevel       string `mapstructure:"log_level"`
	LokiURL        string `mapstructure:"loki_url"`
	TempoEndpoint  string `mapstructure:"tempo_endpoint"`
	MetricsPath    string `mapstructure:"metrics_path"`
	ServiceName    string `mapstructure:"service_name"`
	TraceSampleRatio float64 `mapstructure:"trace_sample_ratio"`
}

type ExternalConfig struct {
	HTTPURL  string `mapstructure:"http_url"`
	GRPCAddr string `mapstructure:"grpc_addr"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		// Allow missing file when env vars / defaults are enough.
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) && !os.IsNotExist(err) && !strings.Contains(err.Error(), "no such file") {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	bindEnv(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}

func LoadDefault() (*Config, error) {
	candidates := []string{
		"configs/config.yaml",
		"config.yaml",
		".env",
	}
	for _, c := range candidates {
		cfg, err := Load(c)
		if err == nil {
			return cfg, nil
		}
	}
	return Load("configs/config.yaml")
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "clean-template")
	v.SetDefault("app.env", "development")
	v.SetDefault("app.version", "0.1.0")
	v.SetDefault("http.port", "8081")
	v.SetDefault("http.read_timeout", "15s")
	v.SetDefault("http.write_timeout", "15s")
	v.SetDefault("http.shutdown_timeout", "10s")
	v.SetDefault("grpc.port", "7000")
	v.SetDefault("grpc.shutdown_timeout", "10s")
	v.SetDefault("postgres.host", "127.0.0.1")
	v.SetDefault("postgres.port", "5432")
	v.SetDefault("postgres.user", "root")
	v.SetDefault("postgres.password", "password123")
	v.SetDefault("postgres.dbname", "sample")
	v.SetDefault("postgres.sslmode", "disable")
	v.SetDefault("postgres.max_open_conns", 25)
	v.SetDefault("postgres.max_idle_conns", 5)
	v.SetDefault("postgres.conn_max_lifetime", "30m")
	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("mongo.uri", "mongodb://127.0.0.1:27017")
	v.SetDefault("mongo.database", "sample")
	v.SetDefault("kafka.brokers", []string{"127.0.0.1:9092"})
	v.SetDefault("kafka.group_id", "clean-template")
	v.SetDefault("kafka.topic", "sample.events")
	v.SetDefault("kafka.dlq_topic", "sample.events.dlq")
	v.SetDefault("kafka.max_handler_retries", 3)
	v.SetDefault("outbox.poll_interval", "1s")
	v.SetDefault("outbox.batch_size", 50)
	v.SetDefault("outbox.max_retries", 5)
	v.SetDefault("outbox.processing_stale_after", "5m")
	v.SetDefault("observability.log_level", "info")
	v.SetDefault("observability.loki_url", "http://127.0.0.1:3100")
	v.SetDefault("observability.tempo_endpoint", "127.0.0.1:4317")
	v.SetDefault("observability.metrics_path", "/metrics")
	v.SetDefault("observability.service_name", "clean-template")
	v.SetDefault("observability.trace_sample_ratio", 1.0)
	v.SetDefault("external.http_url", "https://jsonplaceholder.typicode.com")
	v.SetDefault("external.grpc_addr", "127.0.0.1:7001")
}

func bindEnv(v *viper.Viper) {
	_ = v.BindEnv("app.name", "APP_NAME")
	_ = v.BindEnv("app.secret", "APP_SECRET")
	_ = v.BindEnv("http.port", "PORT", "HTTP_PORT")
	_ = v.BindEnv("grpc.port", "GRPC_PORT")
	_ = v.BindEnv("postgres.host", "DB_HOST")
	_ = v.BindEnv("postgres.port", "DB_PORT")
	_ = v.BindEnv("postgres.user", "DB_USER")
	_ = v.BindEnv("postgres.password", "DB_PASSWORD")
	_ = v.BindEnv("postgres.dbname", "DB_NAME")
	_ = v.BindEnv("redis.addr", "REDIS_ADDR")
	_ = v.BindEnv("redis.password", "REDIS_PASSWORD")
	_ = v.BindEnv("mongo.uri", "MONGO_URI")
	_ = v.BindEnv("mongo.database", "MONGO_DATABASE")
	_ = v.BindEnv("external.http_url", "EXTERNAL_API_URL")
	_ = v.BindEnv("external.grpc_addr", "EXTERNAL_GRPC_ADDR")
	_ = v.BindEnv("observability.loki_url", "LOKI_URL")
	_ = v.BindEnv("observability.tempo_endpoint", "TEMPO_ENDPOINT")
}
