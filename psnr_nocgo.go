//go:build !cgo

package main

// computePSNRCgo falls back to the regular implementation when CGo is not available
func computePSNRCgo(image1Bytes, image2Bytes []byte) (float64, error) {
	return computePSNR(image1Bytes, image2Bytes)
}