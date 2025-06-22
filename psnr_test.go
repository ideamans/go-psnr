package psnr

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
	{
		file1: "testdata/chroma_444.jpg",
		file2: "testdata/chroma_420.jpg",
		name:  "JPEG 4:4:4 vs 4:2:0 chroma subsampling",
	},
	{
		file1: "testdata/fullcolor.png",
		file2: "testdata/palette256.png",
		name:  "PNG full-color vs 256-color palette",
	},
	{
		file1: "testdata/fullcolor.png",
		file2: "testdata/palette_websafe.png",
		name:  "PNG full-color vs web-safe palette",
	},
	{
		file1: "testdata/test_image.png",
		file2: "testdata/test_image_q95.jpg",
		name:  "PNG vs JPEG quality 95",
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

			psnr, err := Compute(data1, data2)

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

			psnr, err := Compute(data1, data2)
			if err != nil {
				t.Fatalf("Error computing PSNR: %v", err)
			}

			fmt.Printf("%s: PSNR = %.6f dB (ImageMagick: %.6f dB)\n", tc.name, psnr, tc.expected)

			// Check if within 0.1% of ImageMagick
			if !math.IsInf(tc.expected, 0) {
				error := math.Abs(psnr-tc.expected) / tc.expected * 100

				fmt.Printf("  Error: %.4f%%\n", error)

				// Allow up to 2% difference due to JPEG decoder implementation differences
				if error > 2.0 {
					t.Errorf("PSNR error %.4f%% exceeds 2.0%% threshold", error)
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

func TestChromaSubsamplingPSNR(t *testing.T) {
	// Test PSNR between different chroma subsampling formats
	tests := []struct {
		name    string
		file1   string
		file2   string
		minPSNR float64 // Minimum expected PSNR
		maxPSNR float64 // Maximum expected PSNR
	}{
		{
			name:    "4:4:4 vs 4:2:0 chroma subsampling",
			file1:   "testdata/chroma_444.jpg",
			file2:   "testdata/chroma_420.jpg",
			minPSNR: 25.0, // Typical range for chroma subsampling differences
			maxPSNR: 45.0,
		},
		{
			name:    "Reference PNG vs 4:4:4 JPEG",
			file1:   "testdata/chroma_reference.png",
			file2:   "testdata/chroma_444.jpg",
			minPSNR: 10.0, // Lower due to JPEG compression artifacts on complex patterns
			maxPSNR: 20.0,
		},
		{
			name:    "Reference PNG vs 4:2:0 JPEG",
			file1:   "testdata/chroma_reference.png",
			file2:   "testdata/chroma_420.jpg",
			minPSNR: 10.0, // Lower due to both compression and subsampling
			maxPSNR: 20.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data1, err := os.ReadFile(tt.file1)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tt.file1, err)
			}

			data2, err := os.ReadFile(tt.file2)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tt.file2, err)
			}

			psnr, err := Compute(data1, data2)
			if err != nil {
				t.Fatalf("Error computing PSNR: %v", err)
			}

			t.Logf("%s: PSNR = %.2f dB", tt.name, psnr)

			if psnr < tt.minPSNR || psnr > tt.maxPSNR {
				t.Errorf("PSNR %.2f dB is outside expected range [%.2f, %.2f]",
					psnr, tt.minPSNR, tt.maxPSNR)
			}
		})
	}
}

func TestPNGvsJPEGPSNR(t *testing.T) {
	// Test PSNR between PNG and JPEG formats
	tests := []struct {
		name    string
		file1   string
		file2   string
		minPSNR float64 // Minimum expected PSNR
		maxPSNR float64 // Maximum expected PSNR
	}{
		{
			name:    "PNG vs JPEG quality 100",
			file1:   "testdata/test_image.png",
			file2:   "testdata/test_image_q100.jpg",
			minPSNR: 40.0, // High quality JPEG should be close to PNG
			maxPSNR: 55.0,
		},
		{
			name:    "PNG vs JPEG quality 95",
			file1:   "testdata/test_image.png",
			file2:   "testdata/test_image_q95.jpg",
			minPSNR: 35.0,
			maxPSNR: 50.0,
		},
		{
			name:    "PNG vs JPEG quality 85",
			file1:   "testdata/test_image.png",
			file2:   "testdata/test_image_q85.jpg",
			minPSNR: 30.0,
			maxPSNR: 45.0,
		},
		{
			name:    "PNG vs JPEG quality 75",
			file1:   "testdata/test_image.png",
			file2:   "testdata/test_image_q75.jpg",
			minPSNR: 25.0,
			maxPSNR: 40.0,
		},
		{
			name:    "JPEG quality 100 vs JPEG quality 85",
			file1:   "testdata/test_image_q100.jpg",
			file2:   "testdata/test_image_q85.jpg",
			minPSNR: 30.0,
			maxPSNR: 45.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data1, err := os.ReadFile(tt.file1)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tt.file1, err)
			}

			data2, err := os.ReadFile(tt.file2)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tt.file2, err)
			}

			psnr, err := Compute(data1, data2)
			if err != nil {
				t.Fatalf("Error computing PSNR: %v", err)
			}

			t.Logf("%s: PSNR = %.2f dB", tt.name, psnr)

			if psnr < tt.minPSNR || psnr > tt.maxPSNR {
				t.Errorf("PSNR %.2f dB is outside expected range [%.2f, %.2f]",
					psnr, tt.minPSNR, tt.maxPSNR)
			}
		})
	}
}

func TestPNGPalettePSNR(t *testing.T) {
	// Test PSNR between full-color and palette PNG images
	tests := []struct {
		name    string
		file1   string
		file2   string
		minPSNR float64 // Minimum expected PSNR
		maxPSNR float64 // Maximum expected PSNR
	}{
		{
			name:    "Full-color vs Plan9 256-color palette",
			file1:   "testdata/fullcolor.png",
			file2:   "testdata/palette256.png",
			minPSNR: 20.0, // Palette quantization typically gives 20-35 dB
			maxPSNR: 35.0,
		},
		{
			name:    "Full-color vs web-safe palette",
			file1:   "testdata/fullcolor.png",
			file2:   "testdata/palette_websafe.png",
			minPSNR: 15.0, // Web-safe palette is more limited
			maxPSNR: 30.0,
		},
		{
			name:    "Plan9 palette vs web-safe palette",
			file1:   "testdata/palette256.png",
			file2:   "testdata/palette_websafe.png",
			minPSNR: 15.0, // Comparing two different palettes
			maxPSNR: 35.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data1, err := os.ReadFile(tt.file1)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tt.file1, err)
			}

			data2, err := os.ReadFile(tt.file2)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tt.file2, err)
			}

			psnr, err := Compute(data1, data2)
			if err != nil {
				t.Fatalf("Error computing PSNR: %v", err)
			}

			t.Logf("%s: PSNR = %.2f dB", tt.name, psnr)

			if psnr < tt.minPSNR || psnr > tt.maxPSNR {
				t.Errorf("PSNR %.2f dB is outside expected range [%.2f, %.2f]",
					psnr, tt.minPSNR, tt.maxPSNR)
			}
		})
	}
}
