package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"chat-transformer/internal/processor"
)

// Build-time variables (set by Makefile)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	var (
		inputFolder  string
		outputFolder string
		showVersion  bool
	)

	// Parse command line arguments
	flag.StringVar(&inputFolder, "i", "", "Input folder path")
	flag.StringVar(&inputFolder, "input", "", "Input folder path")
	flag.StringVar(&inputFolder, "input-folder", "", "Input folder path")
	
	flag.StringVar(&outputFolder, "o", "", "Output folder path")
	flag.StringVar(&outputFolder, "output", "", "Output folder path")
	flag.StringVar(&outputFolder, "output-folder", "", "Output folder path")
	
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information")
	
	flag.Parse()

	// Show version if requested
	if showVersion {
		fmt.Printf("Chat Export Transformer %s\n", Version)
		fmt.Printf("Built: %s\n", BuildTime)
		fmt.Printf("Commit: %s\n", GitCommit)
		return
	}

	// Default paths if not provided
	if inputFolder == "" {
		// Assume we're in the @wisdom folder
		inputFolder = "./raw"
	}
	if outputFolder == "" {
		outputFolder = "./expanded"
	}

	// Convert to absolute paths
	absInput, err := filepath.Abs(inputFolder)
	if err != nil {
		log.Fatalf("Failed to resolve input path: %v", err)
	}

	absOutput, err := filepath.Abs(outputFolder)
	if err != nil {
		log.Fatalf("Failed to resolve output path: %v", err)
	}

	// Verify input folder exists
	if _, err := os.Stat(absInput); os.IsNotExist(err) {
		log.Fatalf("Input folder does not exist: %s", absInput)
	}

	// Create output folder if it doesn't exist
	if err := os.MkdirAll(absOutput, 0755); err != nil {
		log.Fatalf("Failed to create output folder: %v", err)
	}

	fmt.Printf("Chat Export Transformer\n")
	fmt.Printf("=======================\n")
	fmt.Printf("Input folder:  %s\n", absInput)
	fmt.Printf("Output folder: %s\n", absOutput)
	fmt.Printf("\nStarting transformation...\n\n")

	// Initialize and run the processor
	proc := processor.New(absInput, absOutput)
	if err := proc.Run(); err != nil {
		log.Fatalf("Transformation failed: %v", err)
	}

	fmt.Println("\nTransformation completed successfully!")
}