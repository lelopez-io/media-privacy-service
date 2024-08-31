# Media Scrubber Service

This command-line tool scrubs metadata from image and video files, providing a simple way to remove potentially sensitive information from media files before sharing them online.

## Purpose

The Media Scrubber Service simplifies the process of removing potentially sensitive metadata from media files before sharing them online. It supports both image and video files, ensuring privacy and security when distributing media.

## Features

-   Processes HEIC, JPG/JPEG image files, and MOV video files
-   Removes all metadata from images, including EXIF data
-   Converts MOV files to MP4 format while removing metadata
-   Maintains image orientation during processing
-   Generates unique filenames for processed files
-   Supports processing of individual files or entire directories
-   Provides options for image-only processing and output directory cleaning

## Project Structure

```
media-scrubber-service/
├── cmd/
│   ├── cli/
│   │   └── main.go
│   └── server/
│       ├── main.go
│       └── main_test.go
├── internal/
│   └── mediaprocessor/
│       └── processor.go
├── .dockerignore
├── Dockerfile
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## Key Components

### CLI Application (cmd/cli/)

-   `main.go`: The entry point for the command-line interface. It handles local file processing.

### Server Application (cmd/server/)

-   `main.go`: The entry point for the server mode (currently not implemented).
-   `main_test.go`: Contains tests for server flag parsing.

### Media Processor (internal/mediaprocessor/)

-   `processor.go`: Contains the core logic for scrubbing metadata from media files. Key functions include:
    -   `ProcessLocalMediaFile`: Processes local media files
    -   `convertToJpg`: Converts images to JPG format while removing metadata
    -   `convertMovToMp4`: Converts MOV files to MP4 while removing metadata

## Installation

1. Ensure you have Go installed (version 1.22.2 or later recommended).
2. Clone this repository:
    ```sh
    git clone https://github.com/lelopez-io/media-scrubber-service.git
    cd media-scrubber-service
    ```
3. Install dependencies:
    ```sh
    go mod download
    ```

## Usage

### Command-line Flags

-   `--input`: Specify the input directory or file (default is "input")
-   `--output`: Specify the output directory (default is "output")
-   `--clean`: Clean the output directory before processing
-   `--image`: Process only image files

### Examples

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

## Dependencies

-   `github.com/adrium/goheif`: HEIC image processing
-   `github.com/evanoberholster/imagemeta`: Image metadata handling
-   `golang.org/x/image`: Image processing utilities

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License. See the LICENSE file for full details.
