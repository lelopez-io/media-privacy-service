# Media Privacy Service

This tool scrubs metadata from image and video files, providing a simple way to remove potentially sensitive information from media files before sharing them online. It offers both command-line and web interfaces for file processing.

## Purpose

The Media Privacy Service simplifies the process of removing potentially sensitive metadata from media files before sharing them online. It supports both image and video files, ensuring privacy and security when distributing media.

## Features

-   Processes HEIC, JPG/JPEG image files, and MOV/MP4 video files
-   Removes all metadata from images, including EXIF data
-   Converts MOV files to MP4 format while removing metadata
-   Maintains image orientation during processing
-   Generates unique filenames for processed files
-   Supports processing of individual files or entire directories
-   Provides options for image-only processing and output directory cleaning
-   Offers both command-line interface and web interface for file processing
-   Supports concurrent processing for improved performance
-   Provides real-time progress tracking for both CLI and web interface
-   Allows batch downloading of processed files in web interface

## Project Structure

```
media-privacy-service/
├── cmd/
│   ├── cli/
│   │   └── main.go
│   ├── server/
│   │   └── main.go
│   └── webserver/
│       └── main.go
├── internal/
│   └── mediaprocessor/
│       └── processor.go
├── templates/
│   └── index.html
├── Dockerfile
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## Key Components

### CLI Application (cmd/cli/)

-   `main.go`: The entry point for the command-line interface. Handles local file processing with progress tracking.

### Server Application (cmd/server/)

-   `main.go`: Implements the server mode for file processing.

### Web Server Application (cmd/webserver/)

-   `main.go`: Implements the web server, handling file uploads, processing, and downloads through a user-friendly interface.

### Media Processor (internal/mediaprocessor/)

-   `processor.go`: Contains the core logic for scrubbing metadata from media files, including concurrent processing capabilities.

### Web Interface Template (templates/)

-   `index.html`: The HTML template for the web interface, providing drag-and-drop functionality and real-time progress updates.

## Installation

1. Ensure you have Go installed (version 1.22.2 or later recommended).
2. Clone this repository:
    ```sh
    git clone https://github.com/lelopez-io/media-privacy-service.git
    cd media-privacy-service
    ```
3. Install dependencies:
    ```sh
    go mod download
    ```

## Usage

### Command-line Interface

#### Command-line Flags

-   `--input`: Specify the input directory or file (default is "input")
-   `--output`: Specify the output directory (default is "output")
-   `--clean`: Clean the output directory before processing
-   `--image`: Process only image files

#### Examples

Process all supported files in the default input directory:

```sh
go run cmd/cli/main.go
```

Specify custom input and output directories:

```sh
go run cmd/cli/main.go --input=path/to/input --output=path/to/output
```

Process only image files:

```sh
go run cmd/cli/main.go --image
```

Clean the output directory before processing:

```sh
go run cmd/cli/main.go --clean
```

### Local WebServer Mode

1. Start the web server:
    ```sh
    go run cmd/webserver/main.go
    ```

The web interface will be accessible at `http://localhost:8080`. Use it to upload files, monitor processing progress, and download processed files through a user-friendly interface.

### Contained WebServer Mode

To quickly build and run the Docker container with automatic cleanup, use the following command:

```bash
docker run --rm -p 8080:8080 $(docker build -q -t media-privacy-service .)
```

The web interface will be accessible at `http://localhost:8080`.

Note: The Contained WebServer Mode currently needs further optimization. Tests show that processing times can be significantly longer compared to running directly on the host machine, especially for video processing. This is likely due to hardware-specific optimizations that may not be available in the containerized environment.

## Concurrency

The Media Privacy Service utilizes concurrent processing to handle multiple files simultaneously, significantly improving performance for bulk operations. This feature is particularly beneficial when processing a mix of image and video files, as it allows for efficient utilization of system resources.

## Dependencies

-   `github.com/adrium/goheif`: HEIC image processing
-   `github.com/evanoberholster/imagemeta`: Image metadata handling
-   `golang.org/x/image`: Image processing utilities

## License

This project is licensed under the MIT License. See the LICENSE file for full details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Potential Future Enhancements

The following are ideas for potential future enhancements. Feel free to start a discussion on these or propose new ideas via Issues or Pull Requests:

-   User authentication for the web interface
-   Support for additional video formats
-   Enhanced error handling and recovery in concurrent processing
-   Cloud storage integration for processed files
-   Additional tests for the mediaprocessor package

Note: This project currently meets its primary use case. These enhancements are suggestions for those interested in expanding its functionality.
