// Package psnr provides fast PSNR (Peak Signal-to-Noise Ratio) calculation for images.
// It uses integer arithmetic and optimizations while maintaining compatibility
// with ImageMagick's PSNR calculations (typically within 2% margin).
package psnr

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
)

// ComputeFiles calculates PSNR between two image files.
func ComputeFiles(path1, path2 string) (float64, error) {
	data1, err := os.ReadFile(path1)
	if err != nil {
		return 0, fmt.Errorf("failed to read %s: %w", path1, err)
	}

	data2, err := os.ReadFile(path2)
	if err != nil {
		return 0, fmt.Errorf("failed to read %s: %w", path2, err)
	}

	return Compute(data1, data2)
}

// Compute calculates PSNR between two images provided as byte slices.
func Compute(image1Bytes, image2Bytes []byte) (float64, error) {
	img1, format1, err := image.Decode(bytes.NewReader(image1Bytes))
	if err != nil {
		return 0, fmt.Errorf("failed to decode first image: %w", err)
	}

	img2, format2, err := image.Decode(bytes.NewReader(image2Bytes))
	if err != nil {
		return 0, fmt.Errorf("failed to decode second image: %w", err)
	}

	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		return 0, fmt.Errorf("images have different dimensions: %dx%d vs %dx%d",
			bounds1.Dx(), bounds1.Dy(), bounds2.Dx(), bounds2.Dy())
	}

	width := bounds1.Dx()
	height := bounds1.Dy()
	totalPixels := width * height

	// Use integer arithmetic for better performance
	var sumSquaredDiff uint64
	channelCount := 3

	// Optimize alpha channel detection by sampling
	hasAlpha := false
	if format1 == "png" || format2 == "png" {
		// Sample every 16th pixel for faster alpha detection
		step := 16
		if width < 64 || height < 64 {
			step = 4 // Use smaller step for small images
		}
		for y := 0; y < height && !hasAlpha; y += step {
			for x := 0; x < width && !hasAlpha; x += step {
				_, _, _, a1 := img1.At(x+bounds1.Min.X, y+bounds1.Min.Y).RGBA()
				_, _, _, a2 := img2.At(x+bounds2.Min.X, y+bounds2.Min.Y).RGBA()
				if a1 != 0xffff || a2 != 0xffff {
					hasAlpha = true
					channelCount = 4
				}
			}
		}
	}

	// Try fast path for common image types
	switch img1Type := img1.(type) {
	case *image.RGBA:
		if img2RGBA, ok := img2.(*image.RGBA); ok {
			// Fast path for RGBA images
			sumSquaredDiff = computeMSERGBA(img1Type, img2RGBA, hasAlpha)
		} else {
			sumSquaredDiff = computeMSEGeneric(img1, img2, bounds1, bounds2, width, height, hasAlpha)
		}
	case *image.NRGBA:
		if img2NRGBA, ok := img2.(*image.NRGBA); ok {
			// Fast path for NRGBA images (common PNG format)
			sumSquaredDiff = computeMSENRGBA(img1Type, img2NRGBA, hasAlpha)
		} else {
			sumSquaredDiff = computeMSEGeneric(img1, img2, bounds1, bounds2, width, height, hasAlpha)
		}
	case *image.YCbCr:
		if img2YCbCr, ok := img2.(*image.YCbCr); ok {
			// Fast path for YCbCr (JPEG) images
			sumSquaredDiff = computeMSEYCbCr(img1Type, img2YCbCr)
		} else {
			sumSquaredDiff = computeMSEGeneric(img1, img2, bounds1, bounds2, width, height, hasAlpha)
		}
	default:
		sumSquaredDiff = computeMSEGeneric(img1, img2, bounds1, bounds2, width, height, hasAlpha)
	}

	// Convert to MSE
	totalSamples := uint64(totalPixels * channelCount)
	if sumSquaredDiff == 0 {
		return math.Inf(1), nil
	}

	mse := float64(sumSquaredDiff) / float64(totalSamples)

	// Note: Different JPEG decoders (Go's image/jpeg vs libjpeg) may produce
	// slightly different RGB values due to implementation differences in:
	// - YCbCr to RGB conversion rounding
	// - IDCT (Inverse Discrete Cosine Transform) algorithms
	// This can result in small PSNR variations (typically < 1-2%)

	// Fast PSNR calculation
	// PSNR = 10 * log10(255^2 / MSE) = 10 * log10(65025 / MSE)
	psnr := 10 * math.Log10(65025.0/mse)
	return psnr, nil
}

// computeMSEGeneric calculates MSE for any image type
func computeMSEGeneric(img1, img2 image.Image, bounds1, bounds2 image.Rectangle, width, height int, hasAlpha bool) uint64 {
	var sumSquaredDiff uint64

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r1, g1, b1, a1 := img1.At(x+bounds1.Min.X, y+bounds1.Min.Y).RGBA()
			r2, g2, b2, a2 := img2.At(x+bounds2.Min.X, y+bounds2.Min.Y).RGBA()

			// RGBA returns values in 16-bit, convert to 8-bit
			r1, g1, b1, a1 = r1>>8, g1>>8, b1>>8, a1>>8
			r2, g2, b2, a2 = r2>>8, g2>>8, b2>>8, a2>>8

			// Use integer arithmetic for differences
			diffR := int32(r1) - int32(r2)
			diffG := int32(g1) - int32(g2)
			diffB := int32(b1) - int32(b2)

			// Accumulate squared differences as integers
			sumSquaredDiff += uint64(diffR*diffR) + uint64(diffG*diffG) + uint64(diffB*diffB)

			if hasAlpha {
				diffA := int32(a1) - int32(a2)
				sumSquaredDiff += uint64(diffA * diffA)
			}
		}
	}

	return sumSquaredDiff
}

// computeMSERGBA performs fast MSE calculation for RGBA images
func computeMSERGBA(img1, img2 *image.RGBA, hasAlpha bool) uint64 {
	var sumSquaredDiff uint64
	pix1 := img1.Pix
	pix2 := img2.Pix

	// Process 4 bytes at a time (RGBA)
	for i := 0; i < len(pix1); i += 4 {
		diffR := int32(pix1[i]) - int32(pix2[i])
		diffG := int32(pix1[i+1]) - int32(pix2[i+1])
		diffB := int32(pix1[i+2]) - int32(pix2[i+2])

		sumSquaredDiff += uint64(diffR*diffR) + uint64(diffG*diffG) + uint64(diffB*diffB)

		if hasAlpha {
			diffA := int32(pix1[i+3]) - int32(pix2[i+3])
			sumSquaredDiff += uint64(diffA * diffA)
		}
	}

	return sumSquaredDiff
}

// computeMSENRGBA performs fast MSE calculation for NRGBA images (non-premultiplied alpha)
func computeMSENRGBA(img1, img2 *image.NRGBA, hasAlpha bool) uint64 {
	var sumSquaredDiff uint64
	pix1 := img1.Pix
	pix2 := img2.Pix

	// Process 4 bytes at a time (NRGBA)
	for i := 0; i < len(pix1); i += 4 {
		diffR := int32(pix1[i]) - int32(pix2[i])
		diffG := int32(pix1[i+1]) - int32(pix2[i+1])
		diffB := int32(pix1[i+2]) - int32(pix2[i+2])

		sumSquaredDiff += uint64(diffR*diffR) + uint64(diffG*diffG) + uint64(diffB*diffB)

		if hasAlpha {
			diffA := int32(pix1[i+3]) - int32(pix2[i+3])
			sumSquaredDiff += uint64(diffA * diffA)
		}
	}

	return sumSquaredDiff
}

// computeMSEYCbCr performs fast MSE calculation for YCbCr (JPEG) images
func computeMSEYCbCr(img1, img2 *image.YCbCr) uint64 {
	var sumSquaredDiff uint64
	bounds := img1.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Convert YCbCr to RGB for both images
			r1, g1, b1 := color.YCbCrToRGB(img1.YCbCrAt(x, y).Y, img1.YCbCrAt(x, y).Cb, img1.YCbCrAt(x, y).Cr)
			r2, g2, b2 := color.YCbCrToRGB(img2.YCbCrAt(x, y).Y, img2.YCbCrAt(x, y).Cb, img2.YCbCrAt(x, y).Cr)

			diffR := int32(r1) - int32(r2)
			diffG := int32(g1) - int32(g2)
			diffB := int32(b1) - int32(b2)

			sumSquaredDiff += uint64(diffR*diffR) + uint64(diffG*diffG) + uint64(diffB*diffB)
		}
	}

	return sumSquaredDiff
}

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
}
