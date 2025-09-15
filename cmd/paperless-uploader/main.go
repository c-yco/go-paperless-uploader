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
`

func main() {
	// Command-line flags
	filePath := flag.String("file", "", "The path to the document to upload")
	watch := flag.Bool("watch", false, "Watch a directory for new files and upload them")
	createConfig := flag.Bool("create-config", false, "Create an example config.yaml file and exit")
	force := flag.Bool("force", false, "Force overwrite of existing config file")
	flag.Parse()

	if *createConfig {
		if _, err := os.Stat("config.yaml"); err == nil && !*force {
			log.Fatal("config.yaml already exists. Use --force to overwrite.")
		}
		if err := os.WriteFile("config.yaml", []byte(exampleConfig), 0644); err != nil {
			log.Fatalf("Failed to write config file: %v", err)
		}
		if *force {
			log.Println("Overwrote existing config.yaml with example configuration.")
		} else {
			log.Println("Created example config.yaml. Please edit it with your details.")
		}
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a new Paperless client
	client := paperless.NewClient(cfg.PaperlessURL, cfg.APIKey)

	// Log the configuration for debugging
	apiKeyForLogging := ""
	if len(cfg.APIKey) > 4 {
		apiKeyForLogging = "..." + cfg.APIKey[len(cfg.APIKey)-4:]
	} else {
		apiKeyForLogging = "(too short to be valid)"
	}
	log.Printf("Loaded configuration: URL=[%s], APIKey=[%s], WatchFolder=[%s]", cfg.PaperlessURL, apiKeyForLogging, cfg.WatchFolder)

	if *watch {
		log.Printf("Watching directory: %s", cfg.WatchFolder)
		watchDirectory(cfg, client)
	} else if *filePath != "" {
		// Upload the document
		fmt.Printf("Uploading %s to Paperless...\n", *filePath)
		if err := client.UploadDocument(*filePath); err != nil {
			log.Fatalf("Failed to upload document: %v", err)
		}
		fmt.Println("Document uploaded successfully!")
	} else {
		log.Fatal("Either the -file flag or the -watch flag is required")
	}
}

func watchDirectory(cfg *config.Config, client *paperless.Client) {
	// Create the watch folder if it doesn't exist
	if _, err := os.Stat(cfg.WatchFolder); os.IsNotExist(err) {
		log.Printf("Watch folder '%s' not found, creating it.", cfg.WatchFolder)
		if err := os.MkdirAll(cfg.WatchFolder, 0755); err != nil {
			log.Fatalf("Failed to create watch folder: %v", err)
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

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
					if err := client.UploadDocument(event.Name); err != nil {
						log.Printf("Failed to upload document %s: %v", event.Name, err)
					} else {
						log.Printf("Successfully uploaded %s", event.Name)
						// Optionally, move or delete the file after upload
						// os.Remove(event.Name)
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
		log.Fatal(err)
	}

	// Also process existing files in the directory
	err = filepath.Walk(cfg.WatchFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if err := client.UploadDocument(path); err != nil {
				log.Printf("Failed to upload existing document %s: %v", path, err)
			} else {
				log.Printf("Successfully uploaded existing file %s", path)
				// Optionally, move or delete the file after upload
				// os.Remove(path)
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Error processing existing files: %v", err)
	}

	<-done
}
