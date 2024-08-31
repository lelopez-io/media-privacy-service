package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"runtime"

	"github.com/lelopez-io/media-scrubber-service/internal/mediaprocessor"
	"github.com/schollz/progressbar/v3"
)

var (
	inputDir  *string
	outputDir *string
	clean     *bool
	imageOnly *bool
	maxCPU    *bool
)

var bar *progressbar.ProgressBar

const (
	colorRed   = "\033[0;31m"
	colorReset = "\033[0m"
)

func init() {
	inputDir = flag.String("input", "", "Input directory or file")
	outputDir = flag.String("output", "", "Output directory")
	clean = flag.Bool("clean", false, "Clean the output directory before processing")
	imageOnly = flag.Bool("image", false, "Process only image files")
	maxCPU = flag.Bool("max", false, "Use maximum CPU cores for processing")
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
		cleanBar := progressbar.NewOptions(-1,
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionSetWidth(15),
			progressbar.OptionSetDescription("[cyan][1/2][reset] Cleaning output directory..."),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "[green]=[reset]",
				SaucerHead:    "[green]>[reset]",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}))
		
		err := cleanOutputDir(*outputDir)
		if err != nil {
			log.Fatalf("Failed to clean output directory: %v", err)
		}
		cleanBar.Finish()
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
		// Process all files in the directory concurrently
		files, err := ioutil.ReadDir(*inputDir)
		if err != nil {
			log.Fatalf("Error reading input directory: %v", err)
		}

		processFilesConcurrently(files, *inputDir, *outputDir)
	} else {
		// Process single file
		bar = progressbar.NewOptions(1,
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(15),
			progressbar.OptionSetDescription("[cyan][1/1][reset] Processing file..."),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "[green]=[reset]",
				SaucerHead:    "[green]>[reset]",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}))
		processFile(*inputDir, *outputDir)
		bar.Finish()
	}

	fmt.Println("Processing complete.")
}

func processFilesConcurrently(files []os.FileInfo, inputDir, outputDir string) {
	numCPU := runtime.NumCPU()
	numWorkers := numCPU / 2
	if *maxCPU {
		numWorkers = numCPU
	}
	sem := make(chan struct{}, numWorkers)
	var wg sync.WaitGroup

	// Filter out .DS_Store files and count valid files
	validFiles := 0
	for _, file := range files {
		if file.Name() != ".DS_Store" && !file.IsDir() {
			validFiles++
		}
	}

	// Initialize progress bar
	bar = progressbar.NewOptions(validFiles*2, // Multiply by 2 for preparation and processing steps
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan][1/2][reset] Processing files..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	fileCounter := 0
	for _, file := range files {
		if file.IsDir() || file.Name() == ".DS_Store" {
			continue // Skip directories and .DS_Store files
		}

		fileCounter++
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		inputPath := filepath.Join(inputDir, file.Name())
		outputFilename := mediaprocessor.GenerateOrderedFilename(fileCounter, filepath.Ext(file.Name()))
		outputPath := filepath.Join(outputDir, outputFilename)

		go func(inputPath, outputPath string) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			err := processFile(inputPath, outputPath)
			if err != nil {
				printColoredMessageLn(colorRed, fmt.Sprintf("Error processing file %s: %v", inputPath, err))
			}
		}(inputPath, outputPath)
	}

	wg.Wait()
	bar.Finish()
}

func processFile(inputPath, outputPath string) error {
	if !mediaprocessor.IsSupported(inputPath) {
		printColoredMessageLn(colorRed, fmt.Sprintf("Skipping unsupported file: %s", inputPath))
		bar.Add(2) // Add 2 steps for unsupported files
		return nil
	}

	// Check if we should process only images and if the current file is an image
	if *imageOnly && !isImageFile(inputPath) {
		printColoredMessageLn(colorRed, fmt.Sprintf("Skipping non-image file: %s", inputPath))
		bar.Add(2) // Add 2 steps for skipped files
		return nil
	}
	
	// Reading file step
	bar.Describe(fmt.Sprintf("Processing files..."))
	bar.Add(1)

	err := mediaprocessor.ProcessLocalMediaFile(inputPath, outputPath)
	
	// Saving output step
	bar.Describe(fmt.Sprintf("Processing files..."))
	bar.Add(1)

	return err
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
