package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
)

var (
	outputDirectory string
	port            int
)

func init() {
	flag.StringVar(&outputDirectory, "o", ".", "Output directory for compressed images")
	flag.IntVar(&port, "p", 8080, "Port for the server to listen on")
	flag.Parse()
}

func downloadImage(url string) (image.Image, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	var img image.Image
	var format string

	// Determine the image format based on content type
	contentType := resp.Header.Get("Content-Type")
	switch {
	case strings.Contains(contentType, "jpeg"):
		img, _, err = image.Decode(resp.Body)
		format = "jpeg"
	case strings.Contains(contentType, "png"):
		img, _, err = image.Decode(resp.Body)
		format = "png"
	case strings.Contains(contentType, "webp"):
		img, err = webp.Decode(resp.Body)
		format = "webp"
	default:
		return nil, "", fmt.Errorf("unsupported image format")
	}

	if err != nil {
		return nil, "", err
	}

	return img, format, nil
}

func compressImage(img image.Image, format, output string, quality int, resolution string) error {
	// Resize the image if resolution is provided
	if resolution != "" {
		size := strings.Split(resolution, "x")
		width, height := parseResolution(size[0], size[1], img.Bounds().Dx(), img.Bounds().Dy())
		img = resize.Resize(width, height, img, resize.Lanczos3)
	}

	// Create the output file in the specified directory
	out, err := os.Create(filepath.Join(outputDirectory, output))
	if err != nil {
		return err
	}
	defer out.Close()

	// Compress and save the image in the specified format
	switch format {
	case "jpeg":
		options := jpeg.Options{Quality: quality}
		err = jpeg.Encode(out, img, &options)
	case "png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		err = encoder.Encode(out, img)
	case "webp":
		options := &webp.Options{Lossless: true}
		err = webp.Encode(out, img, options)
	default:
		return fmt.Errorf("unsupported output format")
	}

	if err != nil {
		return err
	}

	return nil
}

func generateMD5Hash(input string) string {
	hasher := md5.New()
	hasher.Write([]byte(input))
	return hex.EncodeToString(hasher.Sum(nil))
}

func atoi(s string) int {
	result := 0
	for _, c := range s {
		result = result*10 + int(c-'0')
	}
	return result
}

func parseResolution(width, height string, originalWidth, originalHeight int) (uint, uint) {
	var newWidth, newHeight uint

	if width == "auto" && height == "auto" {
		// If both dimensions are "auto," maintain the original size
		newWidth = uint(originalWidth)
		newHeight = uint(originalHeight)
	} else if width == "auto" {
		// If width is "auto," calculate height maintaining the aspect ratio
		ratio := float64(originalWidth) / float64(originalHeight)
		newHeight = uint(atoi(height))
		newWidth = uint(float64(newHeight) * ratio)
	} else if height == "auto" {
		// If height is "auto," calculate width maintaining the aspect ratio
		ratio := float64(originalHeight) / float64(originalWidth)
		newWidth = uint(atoi(width))
		newHeight = uint(float64(newWidth) * ratio)
	} else {
		// Use the provided width and height
		newWidth = uint(atoi(width))
		newHeight = uint(atoi(height))
	}

	return newWidth, newHeight
}

func compressHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	format := r.URL.Query().Get("output")
	quality := r.URL.Query().Get("quality")
	resolution := r.URL.Query().Get("resolution")

	// Concatenate parameters into a single string
	paramsString := fmt.Sprintf("%s-%s-%s-%s", url, format, quality, resolution)

	// Generate MD5 hash from the concatenated parameters
	hash := generateMD5Hash(paramsString)

	// Generate the output filename using the hash and format
	output := fmt.Sprintf("%s.%s", hash, format)

	// Check if the compressed file already exists in the output directory
	filePath := filepath.Join(outputDirectory, output)
	if _, err := os.Stat(filePath); err == nil {
		// File exists, no need to download and compress again

		// Open and send the existing compressed image file
		compressedFile, err := os.Open(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error opening compressed image file: %s", err), http.StatusInternalServerError)
			return
		}
		defer compressedFile.Close()

		// Set the appropriate Content-Type based on the output format
		var contentType string
		switch format {
		case "jpeg":
			contentType = "image/jpeg"
		case "png":
			contentType = "image/png"
		case "webp":
			contentType = "image/webp"
		default:
			http.Error(w, "Unsupported output format", http.StatusInternalServerError)
			return
		}

		// Set the Content-Type header
		w.Header().Set("Content-Type", contentType)

		// Copy the existing compressed image file to the response writer
		_, err = io.Copy(w, compressedFile)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error sending compressed image: %s", err), http.StatusInternalServerError)
			return
		}

		return
	}

	img, imgFormat, err := downloadImage(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error downloading image: %s", err), http.StatusInternalServerError)
		return
	}

	if format == "" {
		format = imgFormat
	}

	err = compressImage(img, format, output, atoi(quality), resolution)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error compressing image: %s", err), http.StatusInternalServerError)
		return
	}

	// Set the appropriate Content-Type based on the output format
	var contentType string
	switch format {
	case "jpeg":
		contentType = "image/jpeg"
	case "png":
		contentType = "image/png"
	case "webp":
		contentType = "image/webp"
	default:
		http.Error(w, "Unsupported output format", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header
	w.Header().Set("Content-Type", contentType)

	// Open and send the compressed image file
	compressedFile, err := os.Open(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error opening compressed image file: %s", err), http.StatusInternalServerError)
		return
	}
	defer compressedFile.Close()

	// Copy the compressed image file to the response writer
	_, err = io.Copy(w, compressedFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error sending compressed image: %s", err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Image compressed and saved to %s\n", filePath)
}



func main() {
	r := mux.NewRouter()
	r.HandleFunc("/compressor", compressHandler).Methods("GET")
	r.HandleFunc("/compressor/{filename}", compressHandler).Methods("GET")

	http.Handle("/", r)

	fmt.Printf("Server is listening on :%d. Output directory: %s\n", port, outputDirectory)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
