package main

import (
	"fmt"
	"math"
	"os"
	"testing"
)

type testCase struct {
	file1    string
	file2    string
	name     string
	expected float64
}

var testCases = []testCase{
	{
		file1: "testdata/test_original.jpg",
		file2: "testdata/test_original.jpg",
		name:  "JPEG identical images",
	},
	{
		file1: "testdata/test_original.jpg",
		file2: "testdata/quality_50.jpg",
		name:  "JPEG original vs quality 50",
	},
	{
		file1: "testdata/test_original.png",
		file2: "testdata/test_original.png",
		name:  "PNG identical images",
	},
	{
		file1: "testdata/test_original.png",
		file2: "testdata/test_quality_85.png",
		name:  "PNG original vs quality 85",
	},
	{
		file1: "testdata/size1.jpg",
		file2: "testdata/size2.jpg",
		name:  "JPEG different sizes",
	},
}

func TestComputePSNR(t *testing.T) {
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

			psnr, err := computePSNR(data1, data2)
			
			if tc.file1 == tc.file2 {
				if err != nil {
					t.Fatalf("Error computing PSNR for identical images: %v", err)
				}
				if !math.IsInf(psnr, 1) {
					t.Errorf("Expected Inf for identical images, got %f", psnr)
				}
			} else if tc.name == "JPEG different sizes" {
				if err == nil {
					t.Errorf("Expected error for different sized images, got PSNR: %f", psnr)
				}
			} else {
				if err != nil {
					t.Fatalf("Error computing PSNR: %v", err)
				}
				fmt.Printf("%s: PSNR = %.6f dB\n", tc.name, psnr)
			}
		})
	}
}

func TestComputePSNRWithImageMagick(t *testing.T) {
	testCasesWithExpected := []testCase{
		{
			file1:    "testdata/test_original.jpg",
			file2:    "testdata/quality_50.jpg",
			name:     "JPEG original vs quality 50",
			expected: 42.518275,
		},
		{
			file1:    "testdata/test_original.png",
			file2:    "testdata/test_quality_85.png",
			name:     "PNG original vs quality 85",
			expected: math.Inf(1),
		},
	}

	for _, tc := range testCasesWithExpected {
		t.Run(tc.name+" vs ImageMagick", func(t *testing.T) {
			data1, err := os.ReadFile(tc.file1)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tc.file1, err)
			}

			data2, err := os.ReadFile(tc.file2)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tc.file2, err)
			}

			psnr, err := computePSNR(data1, data2)
			if err != nil {
				t.Fatalf("Error computing PSNR: %v", err)
			}

			fmt.Printf("%s: PSNR = %.6f dB (ImageMagick: %.6f dB)\n", tc.name, psnr, tc.expected)
			
			// Check if within 0.1% of ImageMagick
			if !math.IsInf(tc.expected, 0) {
				error := math.Abs(psnr-tc.expected) / tc.expected * 100
				
				fmt.Printf("  Error: %.4f%%\n", error)
				
				if error > 0.1 {
					t.Errorf("PSNR error %.4f%% exceeds 0.1%% threshold", error)
				}
			} else {
				// Both should be Inf
				if !math.IsInf(psnr, 0) {
					t.Errorf("Expected Inf, got %.6f", psnr)
				}
			}
		})
	}
}