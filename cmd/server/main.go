package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/lelopez-io/media-privacy-service/internal/mediaprocessor"
)

var (
	port        *int
	fileCounter uint64
)

func init() {
	port = flag.Int("port", 8080, "Port to run the server on")
}

func main() {
	flag.Parse()

	// Create a temporary directory for output if it doesn't exist
	tempOutputDir := filepath.Join(os.TempDir(), "media-privacy-output")
	if _, err := os.Stat(tempOutputDir); os.IsNotExist(err) {
		err = os.MkdirAll(tempOutputDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create temporary output directory: %v", err)
		}
	}

	http.HandleFunc("/scrub-metadata", handleScrubMetadata(tempOutputDir))

	log.Printf("Server is running on http://localhost:%d\n", *port)
	log.Printf("Use the /scrub-metadata endpoint to process files\n")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func handleScrubMetadata(outputDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Generate a unique filename
		order := int(atomic.AddUint64(&fileCounter, 1))
		outputFilename := mediaprocessor.GenerateOrderedFilename(order, filepath.Ext(header.Filename))
		outputPath := filepath.Join(outputDir, outputFilename)

		// Create a temporary file to save the uploaded content
		tempFile, err := os.CreateTemp("", "upload-*"+filepath.Ext(header.Filename))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		// Copy the uploaded file to the temporary file
		_, err = io.Copy(tempFile, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Process the file
		err = mediaprocessor.ProcessLocalMediaFile(tempFile.Name(), outputPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Serve the processed file
		http.ServeFile(w, r, outputPath)
	}
}
