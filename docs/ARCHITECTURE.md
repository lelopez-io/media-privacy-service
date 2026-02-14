# Architecture

## Project Structure

```
media-privacy-service/
├── cmd/
│   ├── cli/             # Command-line interface
│   │   └── main.go      # Entry point, handles local file processing with progress tracking
│   ├── server/          # Server mode
│   │   └── main.go      # Implements server mode for file processing
│   └── webserver/       # Web server with UI
│       └── main.go      # Handles file uploads, processing, and downloads
├── internal/
│   └── mediaprocessor/  # Core processing logic
│       └── processor.go # Metadata scrubbing, concurrent processing
├── templates/
│   └── index.html       # Web interface with drag-and-drop and progress updates
├── Dockerfile
├── mise.toml
├── go.mod
└── README.md
```

## Features

- Processes HEIC, JPG/JPEG, PNG image files, and MOV/MP4 video files
- Removes all metadata from images, including EXIF data
- Converts MOV files to MP4 format while removing metadata
- Maintains image orientation during processing
- Generates unique filenames for processed files
- Supports processing of individual files or entire directories
- Concurrent processing for improved performance
- Real-time progress tracking for both CLI and web interface
- Batch downloading of processed files in web interface

## Concurrency

The Media Privacy Service utilizes concurrent processing to handle multiple files simultaneously, significantly improving performance for bulk operations. This feature is particularly beneficial when processing a mix of image and video files, as it allows for efficient utilization of system resources.

## Dependencies

- `github.com/adrium/goheif`: HEIC image processing
- `github.com/evanoberholster/imagemeta`: Image metadata handling
- `golang.org/x/image`: Image processing utilities
