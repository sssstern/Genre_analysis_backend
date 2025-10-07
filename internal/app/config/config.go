package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string `mapstructure:"service_host"`
	ServicePort int    `mapstructure:"service_port"`
	Minio       MinioConfig
}

type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	UseSSL          bool
}

func NewConfig() (*Config, error) {
	var err error

	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Warn("No .env file found")
	}

	configName := "config"
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")

	// Устанавливаем значения по умолчанию из .env
	viper.SetDefault("minio.endpoint", os.Getenv("MINIO_ENDPOINT"))
	viper.SetDefault("minio.accesskeyid", os.Getenv("MINIO_ACCESS_KEY"))
	viper.SetDefault("minio.secretaccesskey", os.Getenv("MINIO_SECRET_KEY"))
	viper.SetDefault("minio.bucketname", os.Getenv("MINIO_BUCKET_NAME"))
	viper.SetDefault("minio.usessl", os.Getenv("MINIO_USE_SSL") == "true")
	viper.SetDefault("service_host", "localhost")
	viper.SetDefault("service_port", 8082)

	err = viper.ReadInConfig()
	if err != nil {
		log.Warnf("Config file not found: %v", err)
	}

	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	// Валидация конфигурации Minio
	if cfg.Minio.Endpoint == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT is required")
	}
	if cfg.Minio.AccessKeyID == "" {
		return nil, fmt.Errorf("MINIO_ACCESS_KEY is required")
	}
	if cfg.Minio.SecretAccessKey == "" {
		return nil, fmt.Errorf("MINIO_SECRET_KEY is required")
	}
	if cfg.Minio.BucketName == "" {
		return nil, fmt.Errorf("MINIO_BUCKET_NAME is required")
	}

	log.Info("Config parsed successfully")
	log.Infof("Minio endpoint: %s", cfg.Minio.Endpoint)
	log.Infof("Minio bucket: %s", cfg.Minio.BucketName)

	return cfg, nil
}
