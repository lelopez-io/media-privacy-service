package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/lelopez-io/media-scrubber-service/internal/mediaprocessor"
)

var (
	inputDir  *string
	outputDir *string
	clean     *bool
	imageOnly *bool
)

const (
	colorRed   = "\033[0;31m"
	colorReset = "\033[0m"
)

func init() {
	inputDir = flag.String("input", "", "Input directory or file")
	outputDir = flag.String("output", "", "Output directory")
	clean = flag.Bool("clean", false, "Clean the output directory before processing")
	imageOnly = flag.Bool("image", false, "Process only image files")
}

func main() {
	flag.Parse()

	// Set default directories if not provided
	if *inputDir == "" {
		*inputDir = "input"
	}
	if *outputDir == "" {
		*outputDir = "output"
	}

	// Clean the output directory if the --clean flag is set
	if *clean {
		err := cleanOutputDir(*outputDir)
		if err != nil {
			log.Fatalf("Failed to clean output directory: %v", err)
		}
		fmt.Println("Output directory cleaned.")
	}

	// Check if output directory exists, create if it doesn't
	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		err = os.MkdirAll(*outputDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}
	}

	// Check if input is a file or directory
	fileInfo, err := os.Stat(*inputDir)
	if err != nil {
		log.Fatalf("Error accessing input: %v", err)
	}

	if fileInfo.IsDir() {
		// Process all files in the directory
		files, err := ioutil.ReadDir(*inputDir)
		if err != nil {
			log.Fatalf("Error reading input directory: %v", err)
		}

		for _, file := range files {
			if file.IsDir() {
				continue // Skip directories
			}

			inputPath := filepath.Join(*inputDir, file.Name())
			processFile(inputPath, *outputDir)
		}
	} else {
		// Process single file
		processFile(*inputDir, *outputDir)
	}

	fmt.Println("Processing complete.")
}

func processFile(inputPath, outputDir string) {
	if !mediaprocessor.IsSupported(inputPath) {
		printColoredMessageLn(colorRed, fmt.Sprintf("Skipping unsupported file: %s", inputPath))
		return
	}

	// Check if we should process only images and if the current file is an image
	if *imageOnly && !isImageFile(inputPath) {
		printColoredMessageLn(colorRed, fmt.Sprintf("Skipping non-image file: %s", inputPath))
		return
	}

	fmt.Printf("Processing file: %s\n", inputPath)
	err := mediaprocessor.ProcessLocalMediaFile(inputPath, outputDir)
	if err != nil {
		printColoredMessageLn(colorRed, fmt.Sprintf("Error processing file %s: %v", inputPath, err))
	}
}

func cleanOutputDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}

	return nil
}

func printColoredMessageLn(color, message string) {
	fmt.Println(color + message + colorReset)
}

func isImageFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	imageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".heic": true,
	}
	return imageExtensions[ext]
}