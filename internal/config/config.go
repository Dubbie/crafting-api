package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort     string   `mapstructure:"SERVER_PORT"`
	DBHost         string   `mapstructure:"DB_HOST"`
	DBPort         string   `mapstructure:"DB_PORT"`
	DBUser         string   `mapstructure:"DB_USER"`
	DBPassword     string   `mapstructure:"DB_PASSWORD"`
	DBName         string   `mapstructure:"DB_NAME"`
	AllowedOrigins []string `mapstructure:"ALLOWED_ORIGINS"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.01:5173")

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, fmt.Errorf("error reading config file: %w", err)
		}
		fmt.Println("Config file (.env) not found, using defaults and environment variables.")
		err = nil
	}

	// Special handling for comma-separated strings from environment variables
	// Viper doesn't automatically split env vars into slices like it does for config files.
	if originsStr := viper.GetString("ALLOWED_ORIGINS"); originsStr != "" {
		config.AllowedOrigins = strings.Split(originsStr, ",")
	}

	// Now unmarshal everything
	err = viper.Unmarshal(&config)
	if err != nil {
		return Config{}, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Trim whitespace from origins loaded from config/env
	cleanedOrigins := make([]string, 0, len(config.AllowedOrigins))
	for _, origin := range config.AllowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			cleanedOrigins = append(cleanedOrigins, trimmed)
		}
	}
	config.AllowedOrigins = cleanedOrigins

	return
}
