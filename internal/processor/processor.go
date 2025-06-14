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
	inputPath     string
	outputPath    string
	parser        *parser.Parser
	chatgptParser *parser.ChatGPTParser
	indexer       *indexer.Indexer
	copyMedia     bool
	claudeOnly    bool
	chatgptOnly   bool
}

// New creates a new processor instance
func New(inputPath, outputPath string) *Processor {
	return &Processor{
		inputPath:     inputPath,
		outputPath:    outputPath,
		parser:        parser.New(inputPath),
		chatgptParser: parser.NewChatGPTParser(inputPath),
		indexer:       indexer.New(outputPath),
		copyMedia:     false, // default to not copying media
		claudeOnly:    false,
		chatgptOnly:   false,
	}
}

// SetCopyMedia sets whether to copy media files
func (p *Processor) SetCopyMedia(copy bool) {
	p.copyMedia = copy
}

// SetPlatformModes sets which platforms to process
func (p *Processor) SetPlatformModes(claudeOnly, chatgptOnly bool) {
	p.claudeOnly = claudeOnly
	p.chatgptOnly = chatgptOnly
}

// Run executes the transformation process
func (p *Processor) Run() error {
	fmt.Println("Starting chat export transformation...")

	// Create output directory structure
	if err := p.createDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	var projectStats, claudeStats, chatgptStats ProcessingStats

	// Process Claude exports (unless ChatGPT-only mode)
	if !p.chatgptOnly {
		fmt.Println("Processing Claude projects...")
		var err error
		projectStats, err = p.processClaudeProjects()
		if err != nil {
			fmt.Printf("Warning: Claude project processing failed: %v\n", err)
		} else {
			fmt.Printf("✓ Processed %d Claude projects\n", projectStats.ProjectCount)
		}

		fmt.Println("Processing Claude conversations...")
		claudeStats, err = p.processClaudeConversations()
		if err != nil {
			fmt.Printf("Warning: Claude processing failed: %v\n", err)
		} else {
			fmt.Printf("✓ Processed %d Claude conversations\n", claudeStats.ConversationCount)
		}
	} else {
		fmt.Println("Skipping Claude processing (ChatGPT-only mode)")
	}

	// Process ChatGPT exports (unless Claude-only mode)
	if !p.claudeOnly {
		fmt.Println("Processing ChatGPT conversations...")
		var err error
		chatgptStats, err = p.processChatGPTConversations()
		if err != nil {
			fmt.Printf("Warning: ChatGPT processing failed: %v\n", err)
		} else {
			fmt.Printf("✓ Processed %d ChatGPT conversations\n", chatgptStats.ConversationCount)
		}
	} else {
		fmt.Println("Skipping ChatGPT processing (Claude-only mode)")
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
		ProjectCount:      projectStats.ProjectCount,
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
	ProjectCount      int
	StartTime         time.Time
	EndTime           time.Time
}

// createDirectoryStructure creates the output directory structure
func (p *Processor) createDirectoryStructure() error {
	dirs := []string{
		"claude/projects",
		"claude/chats",
		"claude/media",
		"claude/index",
		"chatgpt/projects",
		"chatgpt/chats",
		"chatgpt/media",
		"chatgpt/index",
		"unified",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(p.outputPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return err
		}
	}

	// Create README files for each container
	if err := p.createREADMEFiles(); err != nil {
		return fmt.Errorf("failed to create README files: %w", err)
	}

	return nil
}

// processClaudeProjects processes Claude project exports
func (p *Processor) processClaudeProjects() (ProcessingStats, error) {
	stats := ProcessingStats{}

	// Load projects
	projects, err := p.parser.ParseClaudeProjects()
	if err != nil {
		return stats, fmt.Errorf("failed to load Claude projects: %w", err)
	}

	// Process each project
	for _, project := range projects {
		// Create project directory
		projectDir := filepath.Join(p.outputPath, "claude", "projects", utils.SanitizeFilename(project.Name))
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			return stats, fmt.Errorf("failed to create project directory: %w", err)
		}

		// Save project metadata
		projectPath := filepath.Join(projectDir, "project.json")
		if err := p.saveProject(project, projectPath); err != nil {
			return stats, fmt.Errorf("failed to save project %s: %w", project.Name, err)
		}

		// Save project documents
		if len(project.Docs) > 0 {
			docsDir := filepath.Join(projectDir, "documents")
			if err := os.MkdirAll(docsDir, 0755); err != nil {
				return stats, fmt.Errorf("failed to create documents directory: %w", err)
			}

			for _, doc := range project.Docs {
				docFilename := utils.SanitizeFilename(doc.Filename) + ".md"
				docPath := filepath.Join(docsDir, docFilename)
				if err := p.saveDocument(doc, docPath); err != nil {
					fmt.Printf("Warning: failed to save document %s: %v\n", doc.Filename, err)
				}
			}
		}

		stats.ProjectCount++
	}

	// Create empty media info for Claude (for consistency)
	claudeMediaInfo := models.ChatGPTMediaInfo{
		Images:             []models.MediaFile{},
		DalleGenerations:   []models.MediaFile{},
		UserUploads:        []models.MediaFile{},
		AudioConversations: []models.AudioConversation{},
	}
	mediaPath := filepath.Join(p.outputPath, "claude", "media", "media_info.json")
	if err := p.saveMediaInfo(claudeMediaInfo, mediaPath); err != nil {
		fmt.Printf("Warning: failed to save Claude media info: %v\n", err)
	}

	return stats, nil
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
			outputDir = filepath.Join(p.outputPath, "claude", "chats", year, month)
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return err
			}
		}

		// Generate filename
		filename := fmt.Sprintf("%s_%s.json",
			conv.Metadata.CreatedDate.Format("2006-01-02"),
			utils.SanitizeFilename(conv.Metadata.Title))
		
		outputPath := filepath.Join(outputDir, filename)
		// Store relative path instead of full path
		relPath, err := filepath.Rel(p.outputPath, outputPath)
		if err != nil {
			// Fallback to constructing relative path manually
			if conv.Metadata.Project != "" {
				relPath = filepath.Join("claude", "projects", utils.SanitizeFilename(conv.Metadata.Project), filename)
			} else {
				year := conv.Metadata.CreatedDate.Format("2006")
				month := conv.Metadata.CreatedDate.Format("01")
				relPath = filepath.Join("claude", "chats", year, month, filename)
			}
		}
		conv.Metadata.FilePath = relPath

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

// processChatGPTConversations processes ChatGPT conversation exports with enhanced parsing
func (p *Processor) processChatGPTConversations() (ProcessingStats, error) {
	stats := ProcessingStats{}

	// Process user info first
	user, err := p.chatgptParser.ParseUserInfo()
	if err != nil {
		fmt.Printf("Warning: failed to parse user info: %v\n", err)
	} else {
		fmt.Printf("Processing ChatGPT export for user: %s\n", user.Name)
	}

	// Process media files
	mediaInfo, err := p.chatgptParser.GetMediaFiles()
	if err != nil {
		fmt.Printf("Warning: failed to scan media files: %v\n", err)
	} else {
		fmt.Printf("Found %d images, %d DALL-E generations, %d user uploads, %d audio conversations\n",
			len(mediaInfo.Images), len(mediaInfo.DalleGenerations), 
			len(mediaInfo.UserUploads), len(mediaInfo.AudioConversations))
		stats.MediaCount = len(mediaInfo.Images) + len(mediaInfo.DalleGenerations) + len(mediaInfo.UserUploads)
	}

	// Save media info and optionally copy files
	if mediaInfo != nil {
		// Convert absolute paths to relative paths from output directory
		relativeMediaInfo := p.convertToRelativePaths(mediaInfo)
		
		// Save media info with relative paths
		mediaPath := filepath.Join(p.outputPath, "chatgpt", "media", "media_info.json")
		if err := p.saveMediaInfo(*relativeMediaInfo, mediaPath); err != nil {
			fmt.Printf("Warning: failed to save media info: %v\n", err)
		}
		
		// Optionally copy media files
		if p.copyMedia {
			fmt.Println("Copying ChatGPT media files...")
			if err := p.copyChatGPTMediaFiles(mediaInfo); err != nil {
				fmt.Printf("Warning: failed to copy some media files: %v\n", err)
			} else {
				fmt.Println("✓ Copied media files")
			}
		}
	}

	// Process conversations using the new parser
	err = p.chatgptParser.ParseConversations(func(chatgpt models.ChatGPTConversation) error {
		conv := parser.ConvertChatGPTToStandard(chatgpt)
		
		// Determine output path
		year := conv.Metadata.CreatedDate.Format("2006")
		month := conv.Metadata.CreatedDate.Format("01")
		outputDir := filepath.Join(p.outputPath, "chatgpt", "chats", year, month)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}

		// Generate filename
		filename := fmt.Sprintf("%s_%s.json",
			conv.Metadata.CreatedDate.Format("2006-01-02"),
			utils.SanitizeFilename(conv.Metadata.Title))
		
		outputPath := filepath.Join(outputDir, filename)
		// Store relative path instead of full path
		relPath, err := filepath.Rel(p.outputPath, outputPath)
		if err != nil {
			relPath = filepath.Join("chatgpt", "chats", year, month, filename) // fallback
		}
		conv.Metadata.FilePath = relPath

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

// saveProject saves a project to disk
func (p *Processor) saveProject(project models.ClaudeProject, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(project)
}

// saveDocument saves a project document to disk as markdown
func (p *Processor) saveDocument(doc models.ClaudeDocument, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write document with metadata header
	content := fmt.Sprintf("# %s\n\n", doc.Filename)
	if doc.CreatedAt != "" {
		content += fmt.Sprintf("**Created:** %s\n\n", doc.CreatedAt)
	}
	content += "---\n\n"
	content += doc.Content

	_, err = file.WriteString(content)
	return err
}

// saveMediaInfo saves media information to disk
func (p *Processor) saveMediaInfo(mediaInfo models.ChatGPTMediaInfo, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(mediaInfo)
}

// generateReport generates a transformation report
func (p *Processor) generateReport(stats ProcessingStats) error {
	report := map[string]interface{}{
		"transformation_completed": time.Now(),
		"statistics": map[string]interface{}{
			"conversations_processed": stats.ConversationCount,
			"messages_processed":      stats.MessageCount,
			"media_files_processed":   stats.MediaCount,
			"projects_processed":      stats.ProjectCount,
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

// createREADMEFiles creates README.md files for each container directory
func (p *Processor) createREADMEFiles() error {
	readmeContents := map[string]string{
		"claude/README.md": `# Claude Export Data

This directory contains processed Claude conversation exports.

## Structure

- **projects/** - Claude projects with their associated documents
- **chats/** - General chat conversations organized by year/month
- **media/** - Media file references and metadata
- **index/** - Search indexes for all Claude conversations
`,
		"claude/projects/README.md": `# Claude Projects

This directory contains Claude projects with their associated documents.

Each project folder contains:
- **project.json** - Project metadata and configuration
- **documents/** - Project-specific documents in markdown format

Projects are organized by project name with sanitized folder names.
`,
		"claude/chats/README.md": `# Claude Chats

This directory contains general Claude chat conversations.

Conversations are organized by:
- **Year/** (e.g., 2024/)
  - **Month/** (e.g., 01/, 02/, ... 12/)
    - Individual conversation JSON files

File naming format: YYYY-MM-DD_ConversationTitle.json
`,
		"claude/media/README.md": `# Claude Media

This directory contains media file references and metadata for Claude conversations.

Currently, Claude exports do not include separate media files, but this structure
is maintained for consistency and future compatibility.
`,
		"claude/index/README.md": `# Claude Search Indexes

This directory contains search indexes for Claude conversations.

- **conversations_index.json** - Master index of all conversations with metadata
`,
		"chatgpt/README.md": `# ChatGPT Export Data

This directory contains processed ChatGPT conversation exports.

## Structure

- **projects/** - Placeholder for consistency (ChatGPT doesn't have projects)
- **chats/** - Chat conversations organized by year/month
- **media/** - Media file references including images, DALL-E generations, and audio
- **index/** - Search indexes for all ChatGPT conversations
`,
		"chatgpt/projects/README.md": `# ChatGPT Projects

This directory is maintained for structural consistency with Claude exports.

**Note:** ChatGPT does not currently support projects. This directory remains empty
but is included to maintain a consistent structure across both platforms.

If ChatGPT adds project support in the future, projects will be organized here
following the same structure as Claude projects.
`,
		"chatgpt/chats/README.md": `# ChatGPT Chats

This directory contains ChatGPT chat conversations.

Conversations are organized by:
- **Year/** (e.g., 2024/)
  - **Month/** (e.g., 01/, 02/, ... 12/)
    - Individual conversation JSON files

File naming format: YYYY-MM-DD_ConversationTitle.json

Note: ChatGPT conversations may include references to media files stored in the
media directory.
`,
		"chatgpt/media/README.md": `# ChatGPT Media

This directory contains media file references and metadata for ChatGPT conversations.

- **media_info.json** - Comprehensive catalog of all media files including:
  - Images uploaded to conversations
  - DALL-E generated images
  - User uploads
  - Audio conversation files

Media files are referenced by their original filenames and paths from the export.
`,
		"chatgpt/index/README.md": `# ChatGPT Search Indexes

This directory contains search indexes for ChatGPT conversations.

- **conversations_index.json** - Master index of all conversations with metadata
`,
		"unified/README.md": `# Unified Search and Analysis

This directory contains cross-platform unified data for searching and analysis
across both Claude and ChatGPT conversations.

Future additions may include:
- Combined search indexes
- Cross-platform analytics
- Merged conversation timelines
- Topic clustering across platforms
`,
	}

	for path, content := range readmeContents {
		fullPath := filepath.Join(p.outputPath, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			// Don't fail if README already exists
			if !os.IsExist(err) {
				return fmt.Errorf("failed to create %s: %w", path, err)
			}
		}
	}

	return nil
}

// convertToRelativePaths converts absolute media file paths to relative paths from output directory
func (p *Processor) convertToRelativePaths(mediaInfo *models.ChatGPTMediaInfo) *models.ChatGPTMediaInfo {
	result := &models.ChatGPTMediaInfo{
		Images:             make([]models.MediaFile, len(mediaInfo.Images)),
		DalleGenerations:   make([]models.MediaFile, len(mediaInfo.DalleGenerations)),
		UserUploads:        make([]models.MediaFile, len(mediaInfo.UserUploads)),
		AudioConversations: make([]models.AudioConversation, len(mediaInfo.AudioConversations)),
	}

	// Convert images
	for i, file := range mediaInfo.Images {
		result.Images[i] = models.MediaFile{
			Name:     file.Name,
			Path:     p.getRelativeMediaPath(file.Path),
			Size:     file.Size,
			Modified: file.Modified,
		}
	}

	// Convert DALL-E generations
	for i, file := range mediaInfo.DalleGenerations {
		result.DalleGenerations[i] = models.MediaFile{
			Name:     file.Name,
			Path:     p.getRelativeMediaPath(file.Path),
			Size:     file.Size,
			Modified: file.Modified,
		}
	}

	// Convert user uploads
	for i, file := range mediaInfo.UserUploads {
		result.UserUploads[i] = models.MediaFile{
			Name:     file.Name,
			Path:     p.getRelativeMediaPath(file.Path),
			Size:     file.Size,
			Modified: file.Modified,
		}
	}

	// Convert audio conversations
	for i, audioConv := range mediaInfo.AudioConversations {
		result.AudioConversations[i] = models.AudioConversation{
			ConversationID: audioConv.ConversationID,
			AudioFiles:     make([]models.MediaFile, len(audioConv.AudioFiles)),
		}
		for j, file := range audioConv.AudioFiles {
			result.AudioConversations[i].AudioFiles[j] = models.MediaFile{
				Name:     file.Name,
				Path:     p.getRelativeMediaPath(file.Path),
				Size:     file.Size,
				Modified: file.Modified,
			}
		}
	}

	return result
}

// getRelativeMediaPath converts an absolute media path to a relative path from output directory
func (p *Processor) getRelativeMediaPath(absolutePath string) string {
	// Try to create a relative path from output directory to the media file
	relPath, err := filepath.Rel(p.outputPath, absolutePath)
	if err != nil {
		// If that fails, create a relative path to the raw directory
		inputBase := filepath.Dir(p.inputPath)
		relToInput, err2 := filepath.Rel(inputBase, absolutePath)
		if err2 != nil {
			// Last resort: return the filename only
			return filepath.Base(absolutePath)
		}
		// Return path relative to the common parent (usually ../raw/...)
		return filepath.Join("..", relToInput)
	}
	return relPath
}