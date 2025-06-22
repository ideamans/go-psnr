package main

import (
	"os"
	"testing"
)

// Benchmark comparison between Go and CGo implementations

func BenchmarkComputePSNR_Go_JPEG(b *testing.B) {
	data1, err := os.ReadFile("testdata/test_original.jpg")
	if err != nil {
		b.Fatalf("Failed to read test_original.jpg: %v", err)
	}
	
	data2, err := os.ReadFile("testdata/quality_50.jpg")
	if err != nil {
		b.Fatalf("Failed to read quality_50.jpg: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := computePSNR(data1, data2)
		if err != nil {
			b.Fatalf("Error computing PSNR: %v", err)
		}
	}
}

func BenchmarkComputePSNR_CGo_JPEG(b *testing.B) {
	data1, err := os.ReadFile("testdata/test_original.jpg")
	if err != nil {
		b.Fatalf("Failed to read test_original.jpg: %v", err)
	}
	
	data2, err := os.ReadFile("testdata/quality_50.jpg")
	if err != nil {
		b.Fatalf("Failed to read quality_50.jpg: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := computePSNRCgo(data1, data2)
		if err != nil {
			b.Fatalf("Error computing PSNR: %v", err)
		}
	}
}

func BenchmarkComputePSNR_Go_PNG(b *testing.B) {
	data1, err := os.ReadFile("testdata/fullcolor.png")
	if err != nil {
		b.Fatalf("Failed to read fullcolor.png: %v", err)
	}
	
	data2, err := os.ReadFile("testdata/palette256.png")
	if err != nil {
		b.Fatalf("Failed to read palette256.png: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := computePSNR(data1, data2)
		if err != nil {
			b.Fatalf("Error computing PSNR: %v", err)
		}
	}
}

func BenchmarkComputePSNR_CGo_PNG(b *testing.B) {
	data1, err := os.ReadFile("testdata/fullcolor.png")
	if err != nil {
		b.Fatalf("Failed to read fullcolor.png: %v", err)
	}
	
	data2, err := os.ReadFile("testdata/palette256.png")
	if err != nil {
		b.Fatalf("Failed to read palette256.png: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := computePSNRCgo(data1, data2)
		if err != nil {
			b.Fatalf("Error computing PSNR: %v", err)
		}
	}
}

// Test large image performance
func BenchmarkComputePSNR_Go_ChromaSubsampling(b *testing.B) {
	data1, err := os.ReadFile("testdata/chroma_444.jpg")
	if err != nil {
		b.Fatalf("Failed to read chroma_444.jpg: %v", err)
	}
	
	data2, err := os.ReadFile("testdata/chroma_420.jpg")
	if err != nil {
		b.Fatalf("Failed to read chroma_420.jpg: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := computePSNR(data1, data2)
		if err != nil {
			b.Fatalf("Error computing PSNR: %v", err)
		}
	}
}

func BenchmarkComputePSNR_CGo_ChromaSubsampling(b *testing.B) {
	data1, err := os.ReadFile("testdata/chroma_444.jpg")
	if err != nil {
		b.Fatalf("Failed to read chroma_444.jpg: %v", err)
	}
	
	data2, err := os.ReadFile("testdata/chroma_420.jpg")
	if err != nil {
		b.Fatalf("Failed to read chroma_420.jpg: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := computePSNRCgo(data1, data2)
		if err != nil {
			b.Fatalf("Error computing PSNR: %v", err)
		}
	}
}

// Benchmark just the MSE calculation part
func BenchmarkMSE_Go_RGBA(b *testing.B) {
	// Create dummy RGBA data
	size := 256 * 256 * 4
	pix1 := make([]uint8, size)
	pix2 := make([]uint8, size)
	
	// Fill with some data
	for i := range pix1 {
		pix1[i] = uint8(i % 256)
		pix2[i] = uint8((i + 10) % 256)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		sum := uint64(0)
		for j := 0; j < len(pix1); j += 4 {
			diffR := int32(pix1[j]) - int32(pix2[j])
			diffG := int32(pix1[j+1]) - int32(pix2[j+1])
			diffB := int32(pix1[j+2]) - int32(pix2[j+2])
			sum += uint64(diffR*diffR) + uint64(diffG*diffG) + uint64(diffB*diffB)
		}
		_ = sum
	}
}

// Benchmark large RGBA images
func BenchmarkComputePSNR_Go_LargeRGBA(b *testing.B) {
	data1, err := os.ReadFile("testdata/large_rgba1.png")
	if err != nil {
		b.Fatalf("Failed to read large_rgba1.png: %v", err)
	}
	
	data2, err := os.ReadFile("testdata/large_rgba2.png")
	if err != nil {
		b.Fatalf("Failed to read large_rgba2.png: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := computePSNR(data1, data2)
		if err != nil {
			b.Fatalf("Error computing PSNR: %v", err)
		}
	}
}

func BenchmarkComputePSNR_CGo_LargeRGBA(b *testing.B) {
	data1, err := os.ReadFile("testdata/large_rgba1.png")
	if err != nil {
		b.Fatalf("Failed to read large_rgba1.png: %v", err)
	}
	
	data2, err := os.ReadFile("testdata/large_rgba2.png")
	if err != nil {
		b.Fatalf("Failed to read large_rgba2.png: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := computePSNRCgo(data1, data2)
		if err != nil {
			b.Fatalf("Error computing PSNR: %v", err)
		}
	}
}

// Benchmark PNG vs JPEG comparison
func BenchmarkComputePSNR_Go_PNGvsJPEG(b *testing.B) {
	data1, err := os.ReadFile("testdata/test_image.png")
	if err != nil {
		b.Fatalf("Failed to read test_image.png: %v", err)
	}
	
	data2, err := os.ReadFile("testdata/test_image_q95.jpg")
	if err != nil {
		b.Fatalf("Failed to read test_image_q95.jpg: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := computePSNR(data1, data2)
		if err != nil {
			b.Fatalf("Error computing PSNR: %v", err)
		}
	}
}

func BenchmarkComputePSNR_CGo_PNGvsJPEG(b *testing.B) {
	data1, err := os.ReadFile("testdata/test_image.png")
	if err != nil {
		b.Fatalf("Failed to read test_image.png: %v", err)
	}
	
	data2, err := os.ReadFile("testdata/test_image_q95.jpg")
	if err != nil {
		b.Fatalf("Failed to read test_image_q95.jpg: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := computePSNRCgo(data1, data2)
		if err != nil {
			b.Fatalf("Error computing PSNR: %v", err)
		}
	}
}