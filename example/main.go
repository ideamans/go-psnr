package main

import (
	"fmt"
	"log"
	"os"

	psnr "github.com/ideamans/go-fast-psnr"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <image1> <image2>\n", os.Args[0])
		os.Exit(1)
	}

	// Calculate PSNR between two image files
	value, err := psnr.ComputeFiles(os.Args[1], os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("PSNR: %.2f dB\n", value)
}
