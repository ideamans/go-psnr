package main

import (
	"os"
	"testing"
)

func BenchmarkComputePSNR(b *testing.B) {
	// Load test images once
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

func BenchmarkComputePSNR_PNG(b *testing.B) {
	// Load test images once
	data1, err := os.ReadFile("testdata/test_original.png")
	if err != nil {
		b.Fatalf("Failed to read test_original.png: %v", err)
	}
	
	data2, err := os.ReadFile("testdata/test_quality_85.png")
	if err != nil {
		b.Fatalf("Failed to read test_quality_85.png: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := computePSNR(data1, data2)
		if err != nil {
			b.Fatalf("Error computing PSNR: %v", err)
		}
	}
}

func BenchmarkComputePSNR_Identical(b *testing.B) {
	// Load test image once
	data, err := os.ReadFile("testdata/test_original.jpg")
	if err != nil {
		b.Fatalf("Failed to read test_original.jpg: %v", err)
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := computePSNR(data, data)
		if err != nil {
			b.Fatalf("Error computing PSNR: %v", err)
		}
	}
}