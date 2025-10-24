package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/c-yco/go-paperless-uploader/internal/config"
	"github.com/c-yco/go-paperless-uploader/pkg/paperless"
	"github.com/fsnotify/fsnotify"
)

const exampleConfig = `paperless_url: "http://localhost:8000"
api_key: "your-api-key"
watch_folder: "consume"
# post_upload_action can be 'delete', 'move', or left empty to do nothing.
post_upload_action: ""
# processed_folder is where files are moved to if post_upload_action is 'move'.
processed_folder: "processed"
# A list of tags to apply to the document.
# tags:
#  - tag1
#  - tag2
`

var (
	logFatal = log.Fatalf //nolint:unused // used in tests
)

func runApp() error {
	// Command-line flags
	filePath := flag.String("file", "", "The path to the document to upload")
	watch := flag.Bool("watch", true, "Watch a directory for new files and upload them")
	createConfig := flag.Bool("create-config", false, "Create an example config.yaml file and exit")
	force := flag.Bool("force", false, "Force overwrite of existing config file")
	flag.Parse()

	if *createConfig {
		if _, err := os.Stat("config.yaml"); err == nil && !*force {
			return fmt.Errorf("config.yaml already exists. Use --force to overwrite")
		}
		if err := os.WriteFile("config.yaml", []byte(exampleConfig), 0644); err != nil {
			return fmt.Errorf("failed to write config file: %v", err)
		}
		if *force {
			log.Println("Overwrote existing config.yaml with example configuration.")
		} else {
			log.Println("Created example config.yaml. Please edit it with your details.")
		}
		return nil
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create a new Paperless client
	client := paperless.NewClient(cfg.PaperlessURL, cfg.APIKey)

	// Get all tags from Paperless
	allTags, err := client.GetTags()
	if err != nil {
		return fmt.Errorf("failed to get tags from Paperless: %v", err)
	}

	// Create a map of tag names to tag IDs for quick lookup
	tagMap := make(map[string]int)
	for _, tag := range allTags {
		tagMap[tag.Name] = tag.ID
	}

	// Convert configured tag names to tag IDs
	var tagIDs []int
	for _, tagName := range cfg.Tags {
		if id, ok := tagMap[tagName]; ok {
			tagIDs = append(tagIDs, id)
		} else {
			log.Printf("Warning: Tag '%s' not found in Paperless and will be ignored.", tagName)
		}
	}

	// Log the configuration for debugging
	apiKeyForLogging := ""
	if len(cfg.APIKey) > 4 {
		apiKeyForLogging = "..." + cfg.APIKey[len(cfg.APIKey)-4:]
	} else {
		apiKeyForLogging = "(too short to be valid)"
	}
	log.Printf("Loaded configuration: URL=[%s], APIKey=[%s], WatchFolder=[%s], PostUploadAction=[%s], ProcessedFolder=[%s], Tags=[%v]", cfg.PaperlessURL, apiKeyForLogging, cfg.WatchFolder, cfg.PostUploadAction, cfg.ProcessedFolder, cfg.Tags)

	if *watch {
		log.Printf("Watching directory: %s", cfg.WatchFolder)
		return watchDirectory(cfg, client, tagIDs)
	} else if *filePath != "" {
		// Upload the document
		fmt.Printf("Uploading %s to Paperless...\n", *filePath)
		if err := client.UploadDocument(*filePath, tagIDs); err != nil {
			return fmt.Errorf("failed to upload document: %v", err)
		}
		fmt.Println("Document uploaded successfully!")
	} else {
		return fmt.Errorf("either the -file flag or the -watch flag is required")
	}
	return nil
}

func watchDirectory(cfg *config.Config, client *paperless.Client, tagIDs []int) error {
	// Create the watch folder if it doesn't exist
	if _, err := os.Stat(cfg.WatchFolder); os.IsNotExist(err) {
		log.Printf("Watch folder '%s' not found, creating it.", cfg.WatchFolder)
		if err := os.MkdirAll(cfg.WatchFolder, 0755); err != nil {
			return fmt.Errorf("failed to create watch folder: %v", err)
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			log.Printf("Error closing watcher: %v", err)
		}
	}()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("New file detected:", event.Name)
					// Wait a moment for the file to be fully written
					time.Sleep(1 * time.Second)
					if err := client.UploadDocument(event.Name, tagIDs); err != nil {
						log.Printf("Failed to upload document %s: %v", event.Name, err)
					} else {
						log.Printf("Successfully uploaded %s", event.Name)
						handlePostUpload(cfg, event.Name)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(cfg.WatchFolder)
	if err != nil {
		return err
	}

	// Also process existing files in the directory
	err = filepath.Walk(cfg.WatchFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if err := client.UploadDocument(path, tagIDs); err != nil {
				log.Printf("Failed to upload existing document %s: %v", path, err)
			} else {
				log.Printf("Successfully uploaded existing file %s", path)
				handlePostUpload(cfg, path)
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Error processing existing files: %v", err)
	}

	<-done
	return nil
}

func handlePostUpload(cfg *config.Config, filePath string) {
	switch cfg.PostUploadAction {
	case "delete":
		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to delete file %s: %v", filePath, err)
		} else {
			log.Printf("Deleted file %s", filePath)
		}
	case "move":
		if _, err := os.Stat(cfg.ProcessedFolder); os.IsNotExist(err) {
			if err := os.MkdirAll(cfg.ProcessedFolder, 0755); err != nil {
				log.Printf("Failed to create processed folder '%s': %v", cfg.ProcessedFolder, err)
				return
			}
		}
		newPath := filepath.Join(cfg.ProcessedFolder, filepath.Base(filePath))
		if err := os.Rename(filePath, newPath); err != nil {
			log.Printf("Failed to move file %s to %s: %v", filePath, newPath, err)
		} else {
			log.Printf("Moved file %s to %s", filePath, newPath)
		}
	}
}
