# go-fast-psnr

Fast PSNR (Peak Signal-to-Noise Ratio) calculation for images in pure Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/ideamans/go-fast-psnr.svg)](https://pkg.go.dev/github.com/ideamans/go-fast-psnr)
[![CI](https://github.com/ideamans/go-fast-psnr/actions/workflows/ci.yml/badge.svg)](https://github.com/ideamans/go-fast-psnr/actions/workflows/ci.yml)

## Features

- **Fast**: Uses integer arithmetic and optimized algorithms
- **Compatible**: Results match ImageMagick within 2% margin
- **Pure Go**: No CGo dependencies, runs everywhere Go runs
- **Simple API**: Easy to use with files or byte slices
- **Format Support**: JPEG and PNG formats

## Installation

```bash
go get github.com/ideamans/go-fast-psnr
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/ideamans/go-fast-psnr/psnr"
)

func main() {
    // Calculate PSNR from file paths
    value, err := psnr.ComputeFiles("image1.jpg", "image2.jpg")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("PSNR: %.2f dB\n", value)
}
```

### Using Byte Slices

```go
// Read images into byte slices
data1, _ := os.ReadFile("image1.png")
data2, _ := os.ReadFile("image2.png")

// Calculate PSNR
value, err := psnr.Compute(data1, data2)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("PSNR: %.2f dB\n", value)
```


## Performance

This package uses several optimizations:

- Integer arithmetic for MSE calculation
- Fast paths for common image formats (RGBA, NRGBA, YCbCr)
- Optimized alpha channel detection
- Direct pixel buffer access for supported formats

## ImageMagick Compatibility

This package is designed to produce PSNR values compatible with ImageMagick (using libjpeg), with results typically within 2% of ImageMagick's calculations. The small differences are due to:

- YCbCr to RGB conversion rounding differences
- IDCT (Inverse Discrete Cosine Transform) implementation variations
- Different JPEG decoder implementations (Go's image/jpeg vs libjpeg)

This level of accuracy makes it suitable as a drop-in replacement for ImageMagick PSNR calculations in most applications.

## Requirements

- Go 1.22 or later

## License

MIT License - see LICENSE file for details.