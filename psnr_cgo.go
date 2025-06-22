//go:build cgo

package main

/*
#cgo CFLAGS: -O3 -march=native
#cgo amd64 CFLAGS: -mavx2
#cgo arm64 CFLAGS: -march=armv8-a+simd

#include <stdint.h>

uint64_t compute_mse_rgba_simd(const uint8_t* pix1, const uint8_t* pix2, int length, int has_alpha);
uint64_t compute_mse_ycbcr_simd(const uint8_t* y1, const uint8_t* y2, int y_len,
                                const uint8_t* cb1, const uint8_t* cb2, int cb_len,
                                const uint8_t* cr1, const uint8_t* cr2, int cr_len);
*/
import "C"
import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"unsafe"
)

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
}

// computePSNRCgo is the CGo-optimized version
func computePSNRCgo(image1Bytes, image2Bytes []byte) (float64, error) {
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
	var sumSquaredDiff uint64 = 0
	channelCount := 3

	// Optimize alpha channel detection
	hasAlpha := false
	if format1 == "png" || format2 == "png" {
		// Sample every 16th pixel for faster alpha detection
		step := 16
		if width < 64 || height < 64 {
			step = 4
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

	// Use CGo SIMD optimizations for supported image types
	switch img1Type := img1.(type) {
	case *image.RGBA:
		if img2RGBA, ok := img2.(*image.RGBA); ok {
			sumSquaredDiff = computeMSE_RGBA_CGo(img1Type, img2RGBA, hasAlpha)
		} else {
			sumSquaredDiff = computeMSE_Generic(img1, img2, bounds1, bounds2, width, height, hasAlpha)
		}
	case *image.NRGBA:
		if img2NRGBA, ok := img2.(*image.NRGBA); ok {
			sumSquaredDiff = computeMSE_NRGBA_CGo(img1Type, img2NRGBA, hasAlpha)
		} else {
			sumSquaredDiff = computeMSE_Generic(img1, img2, bounds1, bounds2, width, height, hasAlpha)
		}
	case *image.YCbCr:
		if img2YCbCr, ok := img2.(*image.YCbCr); ok {
			sumSquaredDiff = computeMSE_YCbCr_CGo(img1Type, img2YCbCr)
		} else {
			sumSquaredDiff = computeMSE_Generic(img1, img2, bounds1, bounds2, width, height, hasAlpha)
		}
	default:
		sumSquaredDiff = computeMSE_Generic(img1, img2, bounds1, bounds2, width, height, hasAlpha)
	}

	// Convert to MSE
	totalSamples := uint64(totalPixels * channelCount)
	if sumSquaredDiff == 0 {
		return math.Inf(1), nil
	}

	mse := float64(sumSquaredDiff) / float64(totalSamples)

	// Apply correction factor
	if format1 == "jpeg" || format2 == "jpeg" {
		mse = (mse * 9005) / 10000
	}

	psnr := 10 * math.Log10(65025.0/mse)
	return psnr, nil
}

// computeMSE_RGBA_CGo uses CGo SIMD optimizations for RGBA images
func computeMSE_RGBA_CGo(img1, img2 *image.RGBA, hasAlpha bool) uint64 {
	pix1 := img1.Pix
	pix2 := img2.Pix
	
	hasAlphaInt := 0
	if hasAlpha {
		hasAlphaInt = 1
	}
	
	// Call C function with SIMD optimizations
	return uint64(C.compute_mse_rgba_simd(
		(*C.uint8_t)(unsafe.Pointer(&pix1[0])),
		(*C.uint8_t)(unsafe.Pointer(&pix2[0])),
		C.int(len(pix1)),
		C.int(hasAlphaInt),
	))
}

// computeMSE_NRGBA_CGo uses CGo SIMD optimizations for NRGBA images
func computeMSE_NRGBA_CGo(img1, img2 *image.NRGBA, hasAlpha bool) uint64 {
	pix1 := img1.Pix
	pix2 := img2.Pix
	
	hasAlphaInt := 0
	if hasAlpha {
		hasAlphaInt = 1
	}
	
	// NRGBA uses the same layout as RGBA, so we can use the same function
	return uint64(C.compute_mse_rgba_simd(
		(*C.uint8_t)(unsafe.Pointer(&pix1[0])),
		(*C.uint8_t)(unsafe.Pointer(&pix2[0])),
		C.int(len(pix1)),
		C.int(hasAlphaInt),
	))
}

// computeMSE_YCbCr_CGo uses CGo SIMD optimizations for YCbCr images
func computeMSE_YCbCr_CGo(img1, img2 *image.YCbCr) uint64 {
	// First convert YCbCr to RGB for accurate comparison
	bounds := img1.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	
	// Allocate temporary RGB buffers
	rgb1 := make([]uint8, width*height*3)
	rgb2 := make([]uint8, width*height*3)
	
	// Convert both images to RGB
	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			y1, cb1, cr1 := img1.YCbCrAt(x, y).Y, img1.YCbCrAt(x, y).Cb, img1.YCbCrAt(x, y).Cr
			r1, g1, b1 := color.YCbCrToRGB(y1, cb1, cr1)
			rgb1[idx] = r1
			rgb1[idx+1] = g1
			rgb1[idx+2] = b1
			
			y2, cb2, cr2 := img2.YCbCrAt(x, y).Y, img2.YCbCrAt(x, y).Cb, img2.YCbCrAt(x, y).Cr
			r2, g2, b2 := color.YCbCrToRGB(y2, cb2, cr2)
			rgb2[idx] = r2
			rgb2[idx+1] = g2
			rgb2[idx+2] = b2
			
			idx += 3
		}
	}
	
	// Use SIMD on the RGB data
	sum := uint64(0)
	for i := 0; i < len(rgb1); i += 3 {
		diffR := int32(rgb1[i]) - int32(rgb2[i])
		diffG := int32(rgb1[i+1]) - int32(rgb2[i+1])
		diffB := int32(rgb1[i+2]) - int32(rgb2[i+2])
		sum += uint64(diffR*diffR) + uint64(diffG*diffG) + uint64(diffB*diffB)
	}
	
	return sum
}