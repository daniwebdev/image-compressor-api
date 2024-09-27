# Image Compression Service Documentation

## Overview
This service allows users to download and compress images from provided URLs. It supports different image formats such as JPEG, PNG, and WebP, with optional resizing and quality adjustments. The server listens for HTTP GET requests, processes the images, and returns the compressed image files.

### Features:
- Supports JPEG, PNG, and WebP formats.
- Image resizing based on custom resolution.
- Image quality adjustment for JPEG.
- Caching based on MD5 hash to avoid redundant compression.
- Domain whitelisting to control which sources are allowed for image download.

---

## Installation

### Prerequisites:
- Go 1.18 or later.
- Required Go packages:
  - `github.com/gorilla/mux`
  - `github.com/chai2010/webp`
  - `github.com/nfnt/resize`

### Clone the repository:
```bash
git clone <repo_url>
```

### Install dependencies:
```bash
go get github.com/gorilla/mux
go get github.com/chai2010/webp
go get github.com/nfnt/resize
```

### Build the service:
```bash
go build -o image-compressor
```

---

## Usage

### Command Line Flags
- `-o` (default: `.`): Specifies the output directory for compressed images.
- `-p` (default: `8080`): Specifies the port on which the server listens.
- `-s` (default: `*`): Comma-separated list of allowed domains for downloading images. Use `*` to allow all domains.

Example:
```bash
./image-compressor -o ./compressed-images -p 8080 -s example.com,another.com
```

### Endpoints

#### 1. **GET /optimize**
This endpoint compresses the image from the given URL, applying optional resizing and quality adjustments. If the compressed image already exists, it is served directly without reprocessing.

##### Query Parameters:
- **url** (required): The image URL to download and compress.
- **output** (optional): Output image format (`jpeg`, `png`, `webp`). Defaults to the format of the original image.
- **quality** (optional): JPEG quality (1-100). Only applies to JPEG format. Defaults to 75.
- **resolution** (optional): Desired resolution in the format `widthxheight`. Use `auto` for either width or height to maintain aspect ratio. Example: `800x600`, `auto x 600`.
- **v** (optional): Versioning parameter used to trigger new compression if image parameters have changed.

##### Example:
```
GET /optimize?url=https://example.com/image.jpg&output=webp&quality=80&resolution=800x600&v=1
```

#### 2. **GET /optimize/{filename}**
This endpoint retrieves an existing compressed image by its filename.

##### Example:
```
GET /optimize/abcd1234.jpeg
```

#### 3. **GET /**
A basic health check endpoint that returns `"ok!"`.

##### Example:
```
GET /
```

---

## How It Works

### Image Download
The `downloadImage` function downloads the image from the provided URL, checks its format, and decodes it into an image object. Supported formats are JPEG, PNG, and WebP.

### Compression
The `compressImage` function resizes the image based on the provided resolution and compresses it based on the chosen format. For JPEG, quality adjustments are possible.

### MD5 Hashing
The image processing parameters (URL, output format, quality, resolution, version) are concatenated into a single string and hashed using MD5 to generate a unique filename for the compressed image.

### File Caching
If the compressed image file already exists, the service retrieves it directly from the output directory, avoiding redundant downloads and compressions.

---

## Error Handling
- If the domain of the image URL is not allowed, the service responds with a `403 Forbidden` error.
- If the image format is unsupported or any other error occurs during image processing, the service responds with a `500 Internal Server Error`.

---

## Example Workflow

1. **Download and Compress Image:**
   ```bash
   curl "http://localhost:8080/optimize?url=https://example.com/image.jpg&output=png&resolution=800x600&quality=90&v=1"
   ```

2. **Retrieve Existing Compressed Image:**
   ```bash
   curl "http://localhost:8080/optimize/<md5-hash>.png"
   ```

---

## License
This project is open source and available under the [MIT License](LICENSE).