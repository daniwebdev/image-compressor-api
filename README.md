# Image Compressor API

## Overview

Image Compressor API is a simple HTTP service written in Go that allows you to compress and resize images from a given URL. It supports popular image formats such as JPEG, PNG, and WebP.

## Features

- Image compression and resizing based on provided parameters.
- Automatic determination of the image format based on the URL's content type.
- Support for JPEG, PNG, and WebP output formats.
- Option to specify output quality and resolution.
- Efficient caching: If the compressed image already exists, it is served without re-compression.

## Getting Started

### Prerequisites

- Go (Golang) installed on your machine.
- [mux](https://github.com/gorilla/mux), [nfnt/resize](https://github.com/nfnt/resize), and [chai2010/webp](https://github.com/chai2010/webp) Go packages.

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/your-repo.git
   cd your-repo
   ```

2. Install dependencies:

   ```bash
   go get -u github.com/gorilla/mux
   go get -u github.com/nfnt/resize
   go get -u github.com/chai2010/webp
   ```

3. Build and run the project:

   ```bash
   go run *.go -o ./tmp
   ```

   Alternatively, for a production build:

   ```bash
   go build -o image-compressor
   ./image-compressor -o ./tmp
   ```

### Usage

To compress an image, make a GET request to the `/compressor` endpoint with the following parameters:

- `url`: URL of the image to be compressed.
- `output`: Desired output format (e.g., "jpeg", "png", "webp").
- `quality`: Output quality (0-100, applicable for JPEG).
- `resolution`: Output resolution in the format "widthxheight" (e.g., "1024x720").

Example:

```bash
curl "http://localhost:8080/compressor?url=https://example.com/image.jpg&output=webp&quality=80&resolution=1024x720"
```

## API Endpoints

### `/compressor`

- **Method:** GET
- **Parameters:**
  - `url` (required): URL of the image to be compressed.
  - `output` (optional): Desired output format (e.g., "jpeg", "png", "webp").
  - `quality` (optional): Output quality (0-100, applicable for JPEG).
  - `resolution` (optional): Output resolution in the format "widthxheight" (e.g., "1024x720").

Example:

```bash
curl "http://localhost:8080/compressor?url=https://example.com/image.jpg&output=webp&quality=80&resolution=1024x720"
```

### Custom Port

By default, the server listens on port `8080`. If you wish to use a custom port, you can specify it during the startup of the server using the `-p` flag. For example:

```bash
./image-compressor -o ./tmp -p 8888
```

## License

This project is licensed under the [MIT License](LICENSE).
