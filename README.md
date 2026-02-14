# Media Privacy Service

Scrubs metadata from image and video files before sharing them online. Supports HEIC, JPG/JPEG images and MOV/MP4 videos.

This project uses [mise](https://mise.jdx.dev/) to manage Go versions automatically.

## Quick Start

### System Dependencies

1. **Install mise for version management:**

```bash
brew install mise
```

2. **Add mise to your shell (add to your `~/.zshrc`):**

```bash
eval "$(mise activate zsh)"
```

3. **Restart your shell or source your config:**

```bash
source ~/.zshrc
```

### Project Setup

```bash
mise trust
mise run setup
```

### Usage

**CLI** - process files in `input/` directory:

```bash
go run cmd/cli/main.go
```

Options:
- `--input=path` - custom input directory or file
- `--output=path` - custom output directory
- `--image` - process only images
- `--clean` - clean output directory first

**Web Interface:**

```bash
go run cmd/webserver/main.go
```

Open `http://localhost:8080` - drag and drop files, monitor progress, download results.

**Docker:**

```bash
docker run --rm -p 8080:8080 $(docker build -q -t media-privacy-service .)
```

Note: Container mode may be slower for video processing due to hardware optimizations.

---

## Additional Documentation

- [Architecture](docs/ARCHITECTURE.md) - Project structure, features, and dependencies
- [Contributing](docs/CONTRIBUTING.md) - How to contribute and future enhancement ideas

## License

MIT License - see LICENSE file.
