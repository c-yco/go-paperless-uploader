package main

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/c-yco/go-paperless-uploader/internal/config"
	"github.com/stretchr/testify/assert"
)

// Mocking log.Fatalf
func TestMain(m *testing.M) {
	logFatal = func(format string, args ...interface{}) {
		// do nothing
	}
	os.Exit(m.Run())
}

func setupTest(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "test-main")
	assert.NoError(t, err)

	// Change working directory to temp dir
	originalWd, err := os.Getwd()
	assert.NoError(t, err)
	os.Chdir(tmpDir)

	return tmpDir, func() {
		os.Chdir(originalWd)
		os.RemoveAll(tmpDir)
	}
}

func TestRunApp(t *testing.T) {
	t.Run("create-config flag", func(t *testing.T) {
		tmpDir, cleanup := setupTest(t)
		defer cleanup()

		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test", "-create-config"}

		err := runApp()
		assert.NoError(t, err)

		_, err = os.Stat(filepath.Join(tmpDir, "config.yaml"))
		assert.NoError(t, err)
	})

	t.Run("file upload", func(t *testing.T) {
		_, cleanup := setupTest(t)
		defer cleanup()

		// Create dummy config and file
		configContent := `
paperless_url: "http://localhost:8000"
api_key: "testkey"
`
		err := os.WriteFile("config.yaml", []byte(configContent), 0644)
		assert.NoError(t, err)

		fileContent := "hello world"
		err = os.WriteFile("test.txt", []byte(fileContent), 0644)
		assert.NoError(t, err)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handle both tags endpoint and upload endpoint
			switch r.URL.Path {
			case "/api/tags/":
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"results": []}`))
			case "/api/documents/post_document/":
				w.WriteHeader(http.StatusOK)
			default:
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		// update config with server url
		configContent = strings.Replace(configContent, "http://localhost:8000", server.URL, 1)
		err = os.WriteFile("config.yaml", []byte(configContent), 0644)
		assert.NoError(t, err)

		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"test", "-file", "test.txt", "-watch=false"}

		err = runApp()
		assert.NoError(t, err)
	})
}

func TestHandlePostUpload(t *testing.T) {
	t.Run("delete action", func(t *testing.T) {
		tmpDir, cleanup := setupTest(t)
		defer cleanup()

		filePath := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(filePath, []byte("content"), 0644)
		assert.NoError(t, err)

		cfg := &config.Config{PostUploadAction: "delete"}
		handlePostUpload(cfg, filePath)

		_, err = os.Stat(filePath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("move action", func(t *testing.T) {
		tmpDir, cleanup := setupTest(t)
		defer cleanup()

		filePath := filepath.Join(tmpDir, "test.txt")
		err := os.WriteFile(filePath, []byte("content"), 0644)
		assert.NoError(t, err)

		processedDir := filepath.Join(tmpDir, "processed")
		cfg := &config.Config{PostUploadAction: "move", ProcessedFolder: processedDir}
		handlePostUpload(cfg, filePath)

		_, err = os.Stat(filePath)
		assert.True(t, os.IsNotExist(err))

		_, err = os.Stat(filepath.Join(processedDir, "test.txt"))
		assert.NoError(t, err)
	})
}
