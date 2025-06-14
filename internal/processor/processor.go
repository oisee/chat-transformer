package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"chat-transformer/internal/indexer"
	"chat-transformer/internal/models"
	"chat-transformer/internal/parser"
	"chat-transformer/internal/utils"
)

// Processor handles the main transformation logic
type Processor struct {
	inputPath  string
	outputPath string
	parser     *parser.Parser
	indexer    *indexer.Indexer
}

// New creates a new processor instance
func New(inputPath, outputPath string) *Processor {
	return &Processor{
		inputPath:  inputPath,
		outputPath: outputPath,
		parser:     parser.New(inputPath),
		indexer:    indexer.New(outputPath),
	}
}

// Run executes the transformation process
func (p *Processor) Run() error {
	fmt.Println("Starting chat export transformation...")

	// Create output directory structure
	if err := p.createDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Process Claude exports
	fmt.Println("Processing Claude conversations...")
	claudeStats, err := p.processClaudeConversations()
	if err != nil {
		fmt.Printf("Warning: Claude processing failed: %v\n", err)
	} else {
		fmt.Printf("✓ Processed %d Claude conversations\n", claudeStats.ConversationCount)
	}

	// Process ChatGPT exports
	fmt.Println("Processing ChatGPT conversations...")
	chatgptStats, err := p.processChatGPTConversations()
	if err != nil {
		fmt.Printf("Warning: ChatGPT processing failed: %v\n", err)
	} else {
		fmt.Printf("✓ Processed %d ChatGPT conversations\n", chatgptStats.ConversationCount)
	}

	// Generate indexes
	fmt.Println("Generating search indexes...")
	if err := p.indexer.GenerateIndexes(); err != nil {
		return fmt.Errorf("failed to generate indexes: %w", err)
	}
	fmt.Println("✓ Generated search indexes")

	// Generate report
	totalStats := ProcessingStats{
		ConversationCount: claudeStats.ConversationCount + chatgptStats.ConversationCount,
		MessageCount:      claudeStats.MessageCount + chatgptStats.MessageCount,
		MediaCount:        claudeStats.MediaCount + chatgptStats.MediaCount,
		StartTime:         time.Now(), // This should be set at the beginning
		EndTime:           time.Now(),
	}

	if err := p.generateReport(totalStats); err != nil {
		fmt.Printf("Warning: failed to generate report: %v\n", err)
	}

	return nil
}

// ProcessingStats holds statistics about the transformation
type ProcessingStats struct {
	ConversationCount int
	MessageCount      int
	MediaCount        int
	StartTime         time.Time
	EndTime           time.Time
}

// createDirectoryStructure creates the output directory structure
func (p *Processor) createDirectoryStructure() error {
	dirs := []string{
		"claude/projects",
		"claude/general-chats",
		"claude/index",
		"chatgpt/conversations",
		"chatgpt/index",
		"unified",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(p.outputPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return err
		}
	}

	return nil
}

// processClaudeConversations processes Claude conversation exports
func (p *Processor) processClaudeConversations() (ProcessingStats, error) {
	stats := ProcessingStats{}

	// Load projects first
	projects, err := p.parser.ParseClaudeProjects()
	if err != nil {
		fmt.Printf("Warning: failed to load Claude projects: %v\n", err)
		projects = []models.ClaudeProject{}
	}

	projectMap := make(map[string]models.ClaudeProject)
	for _, project := range projects {
		projectMap[project.UUID] = project
	}

	// Process conversations
	err = p.parser.ParseClaudeConversations(func(claude models.ClaudeConversation) error {
		conv := parser.ConvertClaudeToStandard(claude, projectMap)
		
		// Determine output path
		var outputDir string
		if conv.Metadata.Project != "" {
			outputDir = filepath.Join(p.outputPath, "claude", "projects", utils.SanitizeFilename(conv.Metadata.Project))
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return err
			}
		} else {
			year := conv.Metadata.CreatedDate.Format("2006")
			month := conv.Metadata.CreatedDate.Format("01")
			outputDir = filepath.Join(p.outputPath, "claude", "general-chats", year, month)
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return err
			}
		}

		// Generate filename
		filename := fmt.Sprintf("%s_%s.json",
			conv.Metadata.CreatedDate.Format("2006-01-02"),
			utils.SanitizeFilename(conv.Metadata.Title))
		
		outputPath := filepath.Join(outputDir, filename)
		conv.Metadata.FilePath = outputPath

		// Save conversation
		if err := p.saveConversation(conv, outputPath); err != nil {
			return err
		}

		// Add to indexer
		p.indexer.AddConversation(conv.Metadata)

		stats.ConversationCount++
		stats.MessageCount += len(conv.Messages)

		return nil
	})

	return stats, err
}

// processChatGPTConversations processes ChatGPT conversation exports
func (p *Processor) processChatGPTConversations() (ProcessingStats, error) {
	stats := ProcessingStats{}

	err := p.parser.ParseChatGPTConversations(func(chatgpt models.ChatGPTConversation) error {
		conv := parser.ConvertChatGPTToStandard(chatgpt)
		
		// Determine output path
		year := conv.Metadata.CreatedDate.Format("2006")
		month := conv.Metadata.CreatedDate.Format("01")
		outputDir := filepath.Join(p.outputPath, "chatgpt", "conversations", year, month)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}

		// Generate filename
		filename := fmt.Sprintf("%s_%s.json",
			conv.Metadata.CreatedDate.Format("2006-01-02"),
			utils.SanitizeFilename(conv.Metadata.Title))
		
		outputPath := filepath.Join(outputDir, filename)
		conv.Metadata.FilePath = outputPath

		// Save conversation
		if err := p.saveConversation(conv, outputPath); err != nil {
			return err
		}

		// Add to indexer
		p.indexer.AddConversation(conv.Metadata)

		stats.ConversationCount++
		stats.MessageCount += len(conv.Messages)

		return nil
	})

	return stats, err
}

// saveConversation saves a conversation to disk
func (p *Processor) saveConversation(conv models.Conversation, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(conv)
}

// generateReport generates a transformation report
func (p *Processor) generateReport(stats ProcessingStats) error {
	report := map[string]interface{}{
		"transformation_completed": time.Now(),
		"statistics": map[string]interface{}{
			"conversations_processed": stats.ConversationCount,
			"messages_processed":      stats.MessageCount,
			"media_files_processed":   stats.MediaCount,
			"processing_duration":     stats.EndTime.Sub(stats.StartTime).String(),
		},
		"output_structure": "see README.md for details",
	}

	reportPath := filepath.Join(p.outputPath, "transformation_report.json")
	file, err := os.Create(reportPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}