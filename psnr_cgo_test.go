package main

import (
	"math"
	"os"
	"testing"
)

// TestCGoAccuracy verifies that CGo implementation produces same results as Go implementation
func TestCGoAccuracy(t *testing.T) {
	testCases := []struct {
		name  string
		file1 string
		file2 string
	}{
		{
			name:  "JPEG comparison",
			file1: "testdata/test_original.jpg",
			file2: "testdata/quality_50.jpg",
		},
		{
			name:  "PNG full-color vs palette",
			file1: "testdata/fullcolor.png",
			file2: "testdata/palette256.png",
		},
		{
			name:  "JPEG chroma subsampling",
			file1: "testdata/chroma_444.jpg",
			file2: "testdata/chroma_420.jpg",
		},
		{
			name:  "Identical images",
			file1: "testdata/test_original.jpg",
			file2: "testdata/test_original.jpg",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data1, err := os.ReadFile(tc.file1)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tc.file1, err)
			}
			
			data2, err := os.ReadFile(tc.file2)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tc.file2, err)
			}
			
			// Compute with Go implementation
			psnrGo, err := computePSNR(data1, data2)
			if err != nil {
				t.Fatalf("Go implementation error: %v", err)
			}
			
			// Compute with CGo implementation
			psnrCGo, err := computePSNRCgo(data1, data2)
			if err != nil {
				t.Fatalf("CGo implementation error: %v", err)
			}
			
			// Compare results
			if math.IsInf(psnrGo, 1) && math.IsInf(psnrCGo, 1) {
				// Both are infinity, which is correct for identical images
				return
			}
			
			// Calculate difference
			diff := math.Abs(psnrGo - psnrCGo)
			relError := diff / psnrGo * 100
			
			t.Logf("Go: %.6f dB, CGo: %.6f dB, Difference: %.6f dB (%.4f%%)", 
				psnrGo, psnrCGo, diff, relError)
			
			// Allow up to 0.01% relative error due to different calculation methods
			if relError > 0.01 {
				t.Errorf("CGo result differs by %.4f%% (exceeds 0.01%% threshold)", relError)
			}
		})
	}
}