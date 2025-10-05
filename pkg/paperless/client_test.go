package paperless

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8000", "test_key")
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8000", client.BaseURL)
	assert.Equal(t, "test_key", client.APIKey)
	assert.NotNil(t, client.HTTPClient)
}

func TestGetTags(t *testing.T) {
	t.Run("successful get tags", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/tags/", r.URL.Path)
			assert.Equal(t, "Token test_key", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"results": [{"id": 1, "name": "tag1"}, {"id": 2, "name": "tag2"}]}`)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test_key")
		tags, err := client.GetTags()
		assert.NoError(t, err)
		assert.Len(t, tags, 2)
		assert.Equal(t, "tag1", tags[0].Name)
	})

	t.Run("failed to send request", func(t *testing.T) {
		client := NewClient("http://invalid-url", "test_key")
		_, err := client.GetTags()
		assert.Error(t, err)
	})

	t.Run("non-200 status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test_key")
		_, err := client.GetTags()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get tags: received status code 500")
	})

	t.Run("invalid json response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"results": "invalid"}`)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test_key")
		_, err := client.GetTags()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode tags response")
	})
}

func TestUploadDocument(t *testing.T) {
	// Create a temporary file for testing uploads
	tmpFile, err := os.CreateTemp("", "test-*.pdf")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString("fake PDF content")
	assert.NoError(t, err)
	tmpFile.Close()

	t.Run("successful upload", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/documents/post_document/", r.URL.Path)
			assert.Equal(t, "Token test_key", r.Header.Get("Authorization"))
			assert.True(t, strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data;"))

			err := r.ParseMultipartForm(10 << 20) // 10 MB
			assert.NoError(t, err)

			file, handler, err := r.FormFile("document")
			assert.NoError(t, err)
			defer file.Close()
			assert.Equal(t, filepath.Base(tmpFile.Name()), handler.Filename)

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test_key")
		err := client.UploadDocument(tmpFile.Name(), nil)
		assert.NoError(t, err)
	})

	t.Run("successful upload with tags", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseMultipartForm(10 << 20)
			assert.NoError(t, err)
			tags := r.Form["tags"]
			assert.Equal(t, []string{"1", "2"}, tags)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(server.URL, "test_key")
		err := client.UploadDocument(tmpFile.Name(), []int{1, 2})
		assert.NoError(t, err)
	})

	t.Run("failed to open file", func(t *testing.T) {
		client := NewClient("http://localhost", "test_key")
		err := client.UploadDocument("/non/existent/file.pdf", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open file")
	})

	t.Run("failed to send request", func(t *testing.T) {
		client := NewClient("http://invalid-url", "test_key")
		err := client.UploadDocument(tmpFile.Name(), nil)
		assert.Error(t, err)
	})

	t.Run("non-200 status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Bad request body")
		}))
		defer server.Close()

		client := NewClient(server.URL, "test_key")
		err := client.UploadDocument(tmpFile.Name(), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to upload document: received status code 400, body: Bad request body")
	})
}
