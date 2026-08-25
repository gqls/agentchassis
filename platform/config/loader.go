// FILE: platform/config/loader.go
package config

import (
	"fmt"
	"log"
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

	// DeliveryPort, when set, starts a SECOND HTTP listener carrying only the
	// customer-facing delivery routes (/c/, later /d/). Used by core-manager.
	//
	// IT IS OPT-IN AND ITS DEFAULT IS OFF, deliberately: the authority it grants
	// is a listener that accepts public traffic (the box proxies customer clicks
	// to it over WireGuard), so the unsafe side is ON. Empty means no second
	// listener AND the delivery routes are mounted nowhere at all — the door is
	// shut, not quietly re-opened on the admin port, which is the failure this
	// whole mechanism exists to prevent (RFC_054 Q2, owner ruling 2026-08-25).
	//
	// Production opts in in the overlay, where a reviewer of the DEPLOYMENT can
	// see the decision rather than having to read the binary (CLAUDE.md
	// 2026-08-02 §2).
	DeliveryPort string `mapstructure:"delivery_port"`
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

	// AutomaticEnv is NOT enough on its own for a key that appears in neither
	// the config file nor a default: Unmarshal populates from viper's known key
	// set, so such a key stays empty however the environment is set, silently.
	//
	// server.delivery_port is exactly that shape — it is set only by
	// core-manager's production overlay (SERVICE_SERVER_DELIVERY_PORT) and no
	// service's YAML names it. Without this bind the delivery listener never
	// starts, every customer link 404s at the box, and the overlay reads
	// perfectly correct while doing nothing. Proven by
	// TestDeliveryPortIsReadFromTheEnvironment, which fails without this line.
	//
	// Binding it does nothing to any other service: unset stays empty, which is
	// the listener's OFF default.
	if err := v.BindEnv("server.delivery_port"); err != nil {
		return nil, fmt.Errorf("config: unable to bind server.delivery_port: %w", err)
	}

	var cfg ServiceConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unable to unmarshal config: %w", err)
	}

	return &cfg, nil
}
