package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/lelopez-io/media-scrubber-service/internal/mediaprocessor"
)

var (
	port *int
)

func init() {
	port = flag.Int("port", 8080, "Port to run the server on")
}

func main() {
	flag.Parse()

	// Create a temporary directory for output if it doesn't exist
	tempOutputDir := filepath.Join(os.TempDir(), "media-scrubber-output")
	if _, err := os.Stat(tempOutputDir); os.IsNotExist(err) {
		err = os.MkdirAll(tempOutputDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create temporary output directory: %v", err)
		}
	}

	log.Printf("Starting server on port %d...\n", *port)
	http.HandleFunc("/scrub-metadata", mediaprocessor.HandleScrubMetadata)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
