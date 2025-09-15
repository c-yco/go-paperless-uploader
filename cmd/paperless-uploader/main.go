package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/your-username/go-paperless-uploader/internal/config"
	"github.com/your-username/go-paperless-uploader/pkg/paperless"
)

func main() {
	// Command-line flag for the file to upload
	filePath := flag.String("file", "", "The path to the document to upload")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("The -file flag is required")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a new Paperless client
	client := paperless.NewClient(cfg.PaperlessURL, cfg.APIKey)

	// Upload the document
	fmt.Printf("Uploading %s to Paperless...\n", *filePath)
	if err := client.UploadDocument(*filePath); err != nil {
		log.Fatalf("Failed to upload document: %v", err)
	}

	fmt.Println("Document uploaded successfully!")
}
