package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Reset viper before each test
	viper.Reset()

	t.Run("successful load from file", func(t *testing.T) {
		viper.Reset()
		// Create a temporary config file
		content := `
paperless_url: "http://test.com"
api_key: "test_key"
watch_folder: "/watch"
post_upload_action: "move"
processed_folder: "/processed"
tags:
  - tag1
  - tag2
`
		tmpDir, err := os.MkdirTemp("", "config-test")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		tmpFile := filepath.Join(tmpDir, "config.yaml")
		err = os.WriteFile(tmpFile, []byte(content), 0600)
		assert.NoError(t, err)

		viper.AddConfigPath(tmpDir)

		cfg, err := Load()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "http://test.com", cfg.PaperlessURL)
		assert.Equal(t, "test_key", cfg.APIKey)
		assert.Equal(t, "/watch", cfg.WatchFolder)
		assert.Equal(t, "move", cfg.PostUploadAction)
		assert.Equal(t, "/processed", cfg.ProcessedFolder)
		assert.Equal(t, []string{"tag1", "tag2"}, cfg.Tags)
	})

	t.Run("config file not found uses defaults", func(t *testing.T) {
		viper.Reset()
		// Point to a non-existent file
		viper.SetConfigFile("/tmp/non-existent-config.yaml")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "http://localhost:8000", cfg.PaperlessURL)
		assert.Equal(t, "watch", cfg.WatchFolder)
		assert.Equal(t, "", cfg.PostUploadAction)
		assert.Equal(t, "processed", cfg.ProcessedFolder)
		assert.Nil(t, cfg.Tags)
	})

	t.Run("environment variables override file", func(t *testing.T) {
		viper.Reset()
		// Create a temporary config file
		content := `
paperless_url: "http://file.com"
api_key: "file_key"
`
		tmpDir, err := os.MkdirTemp("", "config-test-env")
		assert.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		tmpFile := filepath.Join(tmpDir, "config.yaml")
		err = os.WriteFile(tmpFile, []byte(content), 0600)
		assert.NoError(t, err)

		viper.AddConfigPath(tmpDir)

		// Set environment variables
		os.Setenv("UPLOADER_PAPERLESS_URL", "http://env.com")
		os.Setenv("UPLOADER_API_KEY", "env_key")
		os.Setenv("UPLOADER_TAGS", "tagA,tagB")
		defer os.Unsetenv("UPLOADER_PAPERLESS_URL")
		defer os.Unsetenv("UPLOADER_API_KEY")
		defer os.Unsetenv("UPLOADER_TAGS")

		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		cfg, err := Load()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "http://env.com", cfg.PaperlessURL)
		assert.Equal(t, "env_key", cfg.APIKey)
		assert.Equal(t, []string{"tagA", "tagB"}, cfg.Tags)
	})

	t.Run("default values", func(t *testing.T) {
		viper.Reset()
		// Ensure no config file is read
		viper.AddConfigPath("/tmp/non-existent-dir")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Equal(t, "http://localhost:8000", cfg.PaperlessURL)
		assert.Equal(t, "", cfg.APIKey)
		assert.Equal(t, "watch", cfg.WatchFolder)
		assert.Equal(t, "", cfg.PostUploadAction)
		assert.Equal(t, "processed", cfg.ProcessedFolder)
		assert.Nil(t, cfg.Tags)
	})
}
