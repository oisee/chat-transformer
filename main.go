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
		inputFolder     string
		outputFolder    string
		showVersion     bool
		copyMedia       bool
		claudeOnly      bool
		chatgptOnly     bool
		renderMarkdown  bool
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
	
	flag.BoolVar(&copyMedia, "copy-media", false, "Copy media files to output directory (default: false, only store references)")
	
	flag.BoolVar(&claudeOnly, "claude", false, "Process only Claude conversations")
	flag.BoolVar(&claudeOnly, "c", false, "Process only Claude conversations")
	
	flag.BoolVar(&chatgptOnly, "chatgpt", false, "Process only ChatGPT conversations")
	flag.BoolVar(&chatgptOnly, "gpt", false, "Process only ChatGPT conversations")
	flag.BoolVar(&chatgptOnly, "g", false, "Process only ChatGPT conversations")
	
	flag.BoolVar(&renderMarkdown, "render-markdown", false, "Render JSON conversations to readable markdown files")
	flag.BoolVar(&renderMarkdown, "md", false, "Render JSON conversations to readable markdown files")
	
	flag.Parse()

	// Show version if requested
	if showVersion {
		fmt.Printf("Chat Export Transformer %s\n", Version)
		fmt.Printf("Built: %s\n", BuildTime)
		fmt.Printf("Commit: %s\n", GitCommit)
		return
	}

	// Validate mutually exclusive flags
	if claudeOnly && chatgptOnly {
		log.Fatalf("Cannot specify both --claude and --chatgpt flags. Choose one platform to process.")
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

	// Determine what to process
	platformMode := "both platforms"
	if claudeOnly {
		platformMode = "Claude only"
	} else if chatgptOnly {
		platformMode = "ChatGPT only"
	}

	fmt.Printf("Chat Export Transformer\n")
	fmt.Printf("=======================\n")
	fmt.Printf("Input folder:     %s\n", absInput)
	fmt.Printf("Output folder:    %s\n", absOutput)
	fmt.Printf("Copy media:       %v\n", copyMedia)
	fmt.Printf("Platform mode:    %s\n", platformMode)
	fmt.Printf("Render markdown:  %v\n", renderMarkdown)
	fmt.Printf("\nStarting transformation...\n\n")

	// Initialize and run the processor
	proc := processor.New(absInput, absOutput)
	proc.SetCopyMedia(copyMedia)
	proc.SetPlatformModes(claudeOnly, chatgptOnly)
	proc.SetRenderMarkdown(renderMarkdown)
	if err := proc.Run(); err != nil {
		log.Fatalf("Transformation failed: %v", err)
	}

	fmt.Println("\nTransformation completed successfully!")
}