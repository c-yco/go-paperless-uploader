package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config stores the application configuration.
type Config struct {
	PaperlessURL     string   `mapstructure:"paperless_url"`
	APIKey           string   `mapstructure:"api_key"`
	WatchFolder      string   `mapstructure:"watch_folder"`
	PostUploadAction string   `mapstructure:"post_upload_action"`
	ProcessedFolder  string   `mapstructure:"processed_folder"`
	Tags             []string `mapstructure:"tags"`
}

// Load loads the configuration from a file and environment variables.
func Load() (*Config, error) {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")                        // look for config in the working directory
	viper.AddConfigPath("/etc/paperless-uploader/") // and in a system-wide directory
	viper.SetEnvPrefix("UPLOADER")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("paperless_url", "http://localhost:8000")
	viper.SetDefault("watch_folder", "watch")
	viper.SetDefault("post_upload_action", "")
	viper.SetDefault("processed_folder", "processed")
	viper.SetDefault("tags", nil)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file not found; ignore error if it's just not there
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
