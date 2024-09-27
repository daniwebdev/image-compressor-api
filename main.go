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
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"net/url"

	"github.com/chai2010/webp"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
)

var (
	outputDirectory string
	port            int
	allowedDomains  string
	contentTypeMap  = map[string]string{
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"webp": "image/webp",
	}
	initOnce sync.Once
)

func initConfig() {
	initOnce.Do(func() {
		flag.StringVar(&outputDirectory, "o", ".", "Output directory for compressed images")
		flag.IntVar(&port, "p", 8080, "Port for the server to listen on")
		flag.StringVar(&allowedDomains, "s", "*", "Allowed domains separated by comma (,)")
		flag.Parse()
	})
}

func isDomainAllowed(urlString string) bool {
	if allowedDomains == "*" {
		return true
	}

	allowedDomainList := strings.Split(allowedDomains, ",")
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return false
	}

	for _, domain := range allowedDomainList {
		if strings.HasSuffix(parsedURL.Hostname(), domain) {
			return true
		}
	}

	return false
}

func downloadImage(urlString string) (image.Image, string, error) {
	resp, err := http.Get(urlString)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	var img image.Image
	var format string

	switch {
	case strings.Contains(contentType, "jpeg"):
		img, format, err = image.Decode(resp.Body)
		format = "jpeg"
	case strings.Contains(contentType, "png"):
		img, format, err = image.Decode(resp.Body)
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
	if resolution != "" {
		size := strings.Split(resolution, "x")
		width, height := parseResolution(size[0], size[1], img.Bounds().Dx(), img.Bounds().Dy())
		img = resize.Resize(width, height, img, resize.Lanczos3)
	}

	out, err := os.Create(filepath.Join(outputDirectory, output))
	if err != nil {
		return err
	}
	defer out.Close()

	switch format {
	case "jpeg":
		err = jpeg.Encode(out, img, &jpeg.Options{Quality: quality})
	case "png":
		err = (&png.Encoder{CompressionLevel: png.BestCompression}).Encode(out, img)
	case "webp":
		err = webp.Encode(out, img, &webp.Options{Lossless: true})
	default:
		return fmt.Errorf("unsupported output format")
	}

	return err
}

func generateMD5Hash(input string) string {
	hasher := md5.New()
	hasher.Write([]byte(input))
	return hex.EncodeToString(hasher.Sum(nil))
}

func parseResolution(width, height string, originalWidth, originalHeight int) (uint, uint) {
	if width == "auto" && height == "auto" {
		return uint(originalWidth), uint(originalHeight)
	} else if width == "auto" {
		newHeight, _ := strconv.Atoi(height)
		return uint(float64(newHeight) * float64(originalWidth) / float64(originalHeight)), uint(newHeight)
	} else if height == "auto" {
		newWidth, _ := strconv.Atoi(width)
		return uint(newWidth), uint(float64(newWidth) * float64(originalHeight) / float64(originalWidth))
	}
	newWidth, _ := strconv.Atoi(width)
	newHeight, _ := strconv.Atoi(height)
	return uint(newWidth), uint(newHeight)
}

func compressHandler(w http.ResponseWriter, r *http.Request) {
	urlString := r.URL.Query().Get("url")
	format := r.URL.Query().Get("output")
	qualityStr := r.URL.Query().Get("quality")
	resolution := r.URL.Query().Get("resolution")
	version := r.URL.Query().Get("v")

	if !isDomainAllowed(urlString) {
		http.Error(w, "URL domain not allowed", http.StatusForbidden)
		return
	}

	paramsString := fmt.Sprintf("%s-%s-%s-%s-%s", urlString, format, qualityStr, resolution, version)
	hash := generateMD5Hash(paramsString)
	output := fmt.Sprintf("%s.%s", hash, format)

	filePath := filepath.Join(outputDirectory, output)
	if _, err := os.Stat(filePath); err == nil {
		sendExistingFile(w, filePath, format)
		return
	}

	img, imgFormat, err := downloadImage(urlString)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error downloading image: %s", err), http.StatusInternalServerError)
		return
	}

	if format == "" {
		format = imgFormat
	}

	quality, _ := strconv.Atoi(qualityStr)
	err = compressImage(img, format, output, quality, resolution)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error compressing image: %s", err), http.StatusInternalServerError)
		return
	}

	sendExistingFile(w, filePath, format)
}

func sendExistingFile(w http.ResponseWriter, filePath, format string) {
	compressedFile, err := os.Open(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error opening compressed image file: %s", err), http.StatusInternalServerError)
		return
	}
	defer compressedFile.Close()

	w.Header().Set("Content-Type", contentTypeMap[format])
	io.Copy(w, compressedFile)
}

func printBanner() {
	banner := `
------------------------------------
Image Optimizer Service

Author: https://github.com/daniwebdev`
	fmt.Println(banner)
	fmt.Printf("Server is running on port: %d\n", port)
	fmt.Printf("Allowed Domains: %s\n", allowedDomains)
	fmt.Printf("Output Directory: %s\n", outputDirectory)
	fmt.Println("------------------------------------")
}


func main() {
	initConfig()

	r := mux.NewRouter()
	r.HandleFunc("/optimize", compressHandler).Methods("GET")
	r.HandleFunc("/optimize/{filename}", compressHandler).Methods("GET")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok!")
	})

	log.Printf("Server is listening on :%d. Output directory: %s\n", port, outputDirectory)

	printBanner()

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}
