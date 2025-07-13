// FILE: platform/config/loader.go
package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// ServiceConfig is the top-level configuration struct for any service
type ServiceConfig struct {
	ServiceInfo    ServiceInfoConfig      `mapstructure:"service_info"`
	Server         ServerConfig           `mapstructure:"server"`
	Logging        LoggingConfig          `mapstructure:"logging"`
	Observability  ObservabilityConfig    `mapstructure:"observability"`
	Infrastructure InfrastructureConfig   `mapstructure:"infrastructure"`
	Custom         map[string]interface{} `mapstructure:"custom"`
}

type ServiceInfoConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

type ObservabilityConfig struct {
	TracingEndpoint string `mapstructure:"tracing_endpoint"`
}

type InfrastructureConfig struct {
	KafkaBrokers      []string            `mapstructure:"kafka_brokers"`
	ClientsDatabase   DatabaseConfig      `mapstructure:"clients_database"`
	TemplatesDatabase DatabaseConfig      `mapstructure:"templates_database"`
	AuthDatabase      DatabaseConfig      `mapstructure:"auth_database"`
	ObjectStorage     ObjectStorageConfig `mapstructure:"object_storage"`
}

type DatabaseConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	User           string `mapstructure:"user"`
	PasswordEnvVar string `mapstructure:"password_env_var"`
	DBName         string `mapstructure:"db_name"`
	SSLMode        string `mapstructure:"sslmode"`
}

type ObjectStorageConfig struct {
	Provider        string `mapstructure:"provider"`
	Endpoint        string `mapstructure:"endpoint"`
	Bucket          string `mapstructure:"bucket"`
	AccessKeyEnvVar string `mapstructure:"access_key_env_var"`
	SecretKeyEnvVar string `mapstructure:"secret_key_env_var"`
}

// Load reads a YAML config file and overrides with environment variables
func Load(path string) (*ServiceConfig, error) {
	v := viper.New()
	v.SetDefault("server.port", "8080")
	v.SetDefault("logging.level", "info")
	v.SetDefault("infrastructure.database.sslmode", "disable")

	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("config: file not found at %s, relying on defaults and environment variables", path)
		} else {
			return nil, fmt.Errorf("config: error reading config file %s: %w", path, err)
		}
	}

	v.SetEnvPrefix("SERVICE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg ServiceConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unable to unmarshal config: %w", err)
	}

	return &cfg, nil
}
