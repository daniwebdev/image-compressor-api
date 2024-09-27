# Image Optimizer Service Documentation

## Overview
The Image Optimizer Service allows users to download and compress images from specified URLs. It supports a variety of image formats, including JPEG, PNG, and WebP, with additional features for resizing and quality adjustment. The service also caches compressed images to avoid redundant processing, improving performance.

### Key Features:
- Supports multiple image formats: JPEG, PNG, WebP.
- Allows custom resizing based on user-provided resolution.
- Adjustable JPEG quality.
- Caching system based on MD5 hashing to avoid duplicate processing.
- Domain whitelisting to control allowed image sources.

---

## Installation

### Prerequisites:
- **Go version**: 1.18 or later.
- Required Go libraries:
  - `github.com/gorilla/mux` for routing.
  - `github.com/chai2010/webp` for WebP image handling.
  - `github.com/nfnt/resize` for resizing functionality.

### Steps:

1. **Clone the repository:**
   ```bash
   git clone <repo_url>
   ```

2. **Install dependencies:**
   ```bash
   go get github.com/gorilla/mux
   go get github.com/chai2010/webp
   go get github.com/nfnt/resize
   ```

3. **Build the application:**
   ```bash
   go build -o image-compressor
   ```

---

## Usage

### Command-Line Options:
- `-o` (default: `.`): Specifies the output directory where compressed images will be saved.
- `-p` (default: `8080`): Defines the port the service listens on.
- `-s` (default: `*`): Comma-separated list of allowed domains for downloading images. Use `*` to allow all domains.

Example:
```bash
./image-compressor -o ./compressed-images -p 8080 -s example.com,another.com
```

### API Endpoints

#### 1. **GET /optimize**
Compresses an image from a provided URL, optionally resizing and adjusting the quality. If the compressed image already exists, the cached version is returned without recompression.

##### Query Parameters:
- **url** (required): The URL of the image to download and compress.
- **output** (optional): The desired output format (`jpeg`, `png`, `webp`). Defaults to the original format of the image.
- **quality** (optional): JPEG quality (1-100). Only applicable to JPEG images. Default is 75.
- **resolution** (optional): Desired image resolution in the format `widthxheight`. Use `auto` for either width or height to maintain aspect ratio. Example: `800x600`, `auto x 600`.
- **v** (optional): Versioning parameter to force new compression if image parameters change.

##### Example:
```
GET /optimize?url=https://example.com/image.jpg&output=webp&quality=80&resolution=800x600&v=1
```

##### Response:
Returns the compressed image in the specified format or the original format if none is specified.

---

#### 2. **GET /optimize/{filename}**
Fetches a previously compressed image by its filename (typically the MD5 hash of the image parameters).

##### Example:
```
GET /optimize/abcd1234.jpeg
```

##### Response:
Returns the requested image file from the output directory.

---

#### 3. **GET /**
A basic health check endpoint to confirm that the server is running. Returns `"ok!"`.

##### Example:
```
GET /
```

##### Response:
```
ok!
```

---

## How It Works

### 1. **Image Download**
The service downloads the image from the specified URL and decodes it into an `image.Image` object. Supported formats include JPEG, PNG, and WebP.

### 2. **Image Compression**
The downloaded image can be resized based on user-specified resolution parameters. For JPEGs, quality adjustment is supported. After processing, the image is saved in the desired format.

### 3. **MD5 Hashing for Caching**
To avoid redundant processing, the service generates an MD5 hash based on the image URL, desired output format, quality, resolution, and versioning parameters. This hash is used to create a unique filename for the compressed image. If the file already exists, the cached version is returned.

### 4. **Caching and Performance**
If the compressed image already exists in the output directory, the service retrieves and returns it directly, bypassing the need to download and process the image again. This caching mechanism significantly improves performance and reduces redundant computation.

---

## Error Handling

- **403 Forbidden:** Returned if the domain of the image URL is not on the allowed list (`-s` parameter).
- **500 Internal Server Error:** Occurs if an unsupported image format is provided or an error occurs during the image processing stage.

---

## Example Workflow

1. **Download and compress an image:**
   ```bash
   curl "http://localhost:8080/optimize?url=https://example.com/image.jpg&output=png&resolution=800x600&quality=90&v=1"
   ```

2. **Retrieve an existing compressed image:**
   ```bash
   curl "http://localhost:8080/optimize/<md5-hash>.png"
   ```

---

## License
This project is licensed under the [MIT License](LICENSE).

---

### Contributions
Feel free to contribute to this project by submitting issues or pull requests on the official repository.