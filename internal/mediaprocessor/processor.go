package mediaprocessor

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adrium/goheif"
	"github.com/evanoberholster/imagemeta"
	"golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
)

// fileCounter is used to generate ordered prefixes for filenames
var fileCounter uint64

// SupportedExtensions is a map of supported file extensions and their corresponding processing functions
var SupportedExtensions = map[string]func(string, string) error{
	".heic": convertToJpg,
	".HEIC": convertToJpg,
	".jpg":  convertToJpg,
	".jpeg": convertToJpg,
	".JPG":  convertToJpg,
	".JPEG": convertToJpg,
	".png":  convertToJpg,
	".PNG":  convertToJpg,
	".mov":  convertMovToMp4,
	".MOV":  convertMovToMp4,
	".mp4":  convertMovToMp4,
	".MP4":  convertMovToMp4,
}

// ProcessLocalMediaFile handles the processing of a single media file
func ProcessLocalMediaFile(inputPath, outputPath string) error {
	ext := strings.ToLower(filepath.Ext(inputPath))

	processFunc, supported := SupportedExtensions[ext]
	if !supported {
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	err := processFunc(inputPath, outputPath)
	if err != nil {
		return fmt.Errorf("error processing file: %v", err)
	}

	fmt.Printf("Processed %s to %s\n", inputPath, outputPath)
	return nil
}

// convertToJpg converts a file to JPG without preserving metadata but maintaining orientation
func convertToJpg(input, output string) error {
	fileInput, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("error opening input file: %v", err)
	}
	defer fileInput.Close()

	// Extract orientation
	orientation := 1
	metadata, err := imagemeta.Decode(fileInput)
	if err == nil && metadata.Orientation > 0 && metadata.Orientation <= 8 {
		orientation = int(metadata.Orientation)
	}

	// Reset file pointer to the beginning
	_, err = fileInput.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error resetting file pointer: %v", err)
	}

	var img image.Image

	// Determine file type and decode accordingly
	ext := strings.ToLower(filepath.Ext(input))
	switch ext {
	case ".heic":
		img, err = goheif.Decode(fileInput)
	case ".png":
		img, err = png.Decode(fileInput)
	default:
		img, _, err = image.Decode(fileInput)
	}

	if err != nil {
		return fmt.Errorf("error decoding image: %v", err)
	}

	// Apply orientation
	img = ApplyOrientation(img, orientation)

	fileOutput, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer fileOutput.Close()

	// Encode as JPEG without any metadata
	opts := jpeg.Options{Quality: 90}
	err = jpeg.Encode(fileOutput, img, &opts)
	if err != nil {
		return fmt.Errorf("error encoding JPEG: %v", err)
	}

	return nil
}

// convertMovToMp4 converts a MOV or MP4 file to MP4 using FFmpeg
func convertMovToMp4(input, output string) error {
	cmd := exec.Command("ffmpeg",
		"-i", input,
		"-map_metadata", "-1", // Remove all metadata
		"-c:v", "libx264",
		"-crf", "23",
		"-preset", "medium",
		"-c:a", "aac",
		"-b:a", "128k",
		"-movflags", "+faststart",
		"-y", output)

	var stderr strings.Builder
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("FFmpeg command failed: %v\nFFmpeg error output:\n%s", err, stderr.String())
	}

	return nil
}

// IsSupported checks if a given file is supported based on its extension
func IsSupported(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	_, supported := SupportedExtensions[ext]
	return supported
}

// GenerateOrderedFilename generates a filename with an ordered prefix
func GenerateOrderedFilename(order int, ext string) string {
	// Generate the ordered prefix
	orderPrefix := fmt.Sprintf("%06d", order)

	// Generate the random part
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		panic(err) // handle error appropriately in production code
	}
	randomPart := hex.EncodeToString(randomBytes)

	// Determine the output extension
	outputExt := ".jpg"
	if strings.ToLower(ext) == ".mov" || strings.ToLower(ext) == ".mp4" {
		outputExt = ".mp4"
	}

	return fmt.Sprintf("%s_%s%s", orderPrefix, randomPart, outputExt)
}

func ApplyOrientation(img image.Image, orientation int) image.Image {
	switch orientation {
	case 1:
		return img // No rotation needed
	case 2:
		return flipHorizontal(img)
	case 3:
		return rotate180(img)
	case 4:
		return flipVertical(img)
	case 5:
		return transpose(img)
	case 6:
		return rotate90(img) // Changed from rotate270 to rotate90
	case 7:
		return transverse(img)
	case 8:
		return rotate270(img)
	default:
		fmt.Printf("Warning: Unknown orientation %d, returning original image\n", orientation)
		return img
	}
}

func flipHorizontal(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			newImg.Set(bounds.Max.X-x-1, y, img.At(x, y))
		}
	}
	return newImg
}

func flipVertical(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			newImg.Set(x, bounds.Max.Y-y-1, img.At(x, y))
		}
	}
	return newImg
}

func transpose(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			newImg.Set(y, x, img.At(x, y))
		}
	}
	return newImg
}

func transverse(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			newImg.Set(bounds.Max.Y-y-1, bounds.Max.X-x-1, img.At(x, y))
		}
	}
	return newImg
}

func rotate90(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))
	draw.NearestNeighbor.Transform(newImg, f64.Aff3{0, -1, float64(bounds.Dy()), 1, 0, 0}, img, bounds, draw.Over, nil)
	return newImg
}

func rotate180(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	draw.NearestNeighbor.Transform(newImg, f64.Aff3{-1, 0, float64(bounds.Dx()), 0, -1, float64(bounds.Dy())}, img, bounds, draw.Over, nil)
	return newImg
}

func rotate270(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))
	draw.NearestNeighbor.Transform(newImg, f64.Aff3{0, 1, 0, -1, 0, float64(bounds.Dx())}, img, bounds, draw.Over, nil)
	return newImg
}

// GetOrientation extracts the orientation from an image file
func GetOrientation(input string) (int, error) {
	fileInput, err := os.Open(input)
	if err != nil {
		return 1, fmt.Errorf("error opening input file: %v", err)
	}
	defer fileInput.Close()

	metadata, err := imagemeta.Decode(fileInput)
	if err != nil {
		return 1, err
	}

	if metadata.Orientation > 0 && metadata.Orientation <= 8 {
		return int(metadata.Orientation), nil
	}

	return 1, fmt.Errorf("invalid orientation value")
}
