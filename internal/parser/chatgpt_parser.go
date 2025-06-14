package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"chat-transformer/internal/models"
)

const (
	// Number of parallel workers for conversation processing
	ConversationWorkers = 25
)

// ChatGPTParser handles parsing of ChatGPT exports with streaming support
type ChatGPTParser struct {
	inputPath string
}

// conversationJob represents a conversation to be processed
type conversationJob struct {
	rawConv models.ChatGPTConversationRaw
	index   int
}

// NewChatGPTParser creates a new ChatGPT parser instance
func NewChatGPTParser(inputPath string) *ChatGPTParser {
	return &ChatGPTParser{
		inputPath: inputPath,
	}
}

// ParseConversations parses ChatGPT conversations.json with streaming support
func (p *ChatGPTParser) ParseConversations(callback func(models.ChatGPTConversation) error) error {
	filePath := filepath.Join(p.inputPath, "chat-gpt-2025-06-13", "conversations.json")
	
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open ChatGPT conversations file: %w", err)
	}
	defer file.Close()

	// Check file size to determine parsing strategy
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	
	fileSize := fileInfo.Size()
	fmt.Printf("ChatGPT conversations.json size: %.2f MB\n", float64(fileSize)/(1024*1024))

	// For very large files (>100MB), use streaming approach
	if fileSize > 100*1024*1024 {
		return p.parseConversationsStreaming(file, callback)
	}
	
	// For smaller files, use standard approach
	return p.parseConversationsStandard(file, callback)
}

// parseConversationsStreaming handles large single-line JSON files
func (p *ChatGPTParser) parseConversationsStreaming(file *os.File, callback func(models.ChatGPTConversation) error) error {
	fmt.Println("Using streaming parser for large ChatGPT file...")
	
	// Read the entire file content (since it's a single line)
	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the JSON array
	var conversations []models.ChatGPTConversationRaw
	if err := json.Unmarshal(content, &conversations); err != nil {
		return fmt.Errorf("failed to parse ChatGPT conversations JSON: %w", err)
	}

	fmt.Printf("Successfully parsed %d ChatGPT conversations\n", len(conversations))

	// Process conversations in parallel
	return p.processConversationsParallel(conversations, callback)
}

// parseConversationsStandard handles normally sized files
func (p *ChatGPTParser) parseConversationsStandard(file *os.File, callback func(models.ChatGPTConversation) error) error {
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read ChatGPT conversations file: %w", err)
	}

	var conversations []models.ChatGPTConversationRaw
	if err := json.Unmarshal(data, &conversations); err != nil {
		return fmt.Errorf("failed to parse ChatGPT conversations JSON: %w", err)
	}

	for i, rawConv := range conversations {
		conv, err := p.convertRawConversation(rawConv)
		if err != nil {
			fmt.Printf("Warning: failed to convert conversation %d: %v\n", i, err)
			continue
		}

		if err := callback(conv); err != nil {
			fmt.Printf("Warning: callback failed for ChatGPT conversation %s: %v\n", conv.ID, err)
		}
	}

	return nil
}

// processConversationsParallel processes conversations using parallel workers
func (p *ChatGPTParser) processConversationsParallel(conversations []models.ChatGPTConversationRaw, callback func(models.ChatGPTConversation) error) error {
	totalConversations := len(conversations)
	if totalConversations == 0 {
		return nil
	}

	// Create channels for job distribution and progress tracking
	jobChan := make(chan conversationJob, 100) // Buffered channel
	resultChan := make(chan error, totalConversations)
	progressChan := make(chan int, totalConversations)

	// Determine number of workers
	numWorkers := ConversationWorkers
	if totalConversations < numWorkers {
		numWorkers = totalConversations
	}

	fmt.Printf("Processing conversations with %d workers...\n", numWorkers)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go p.conversationWorker(&wg, jobChan, resultChan, progressChan, callback)
	}

	// Start progress reporter
	var progressWg sync.WaitGroup
	progressWg.Add(1)
	go p.progressReporter(&progressWg, progressChan, totalConversations)

	// Send jobs to workers
	for i, rawConv := range conversations {
		jobChan <- conversationJob{
			rawConv: rawConv,
			index:   i,
		}
	}
	close(jobChan)

	// Wait for all workers to complete
	wg.Wait()
	close(resultChan)
	close(progressChan)

	// Wait for progress reporter to finish
	progressWg.Wait()

	// Count successful and failed conversions
	successCount := 0
	var errors []error
	for err := range resultChan {
		if err != nil {
			errors = append(errors, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("Successfully processed %d valid conversations\n", successCount)
	if len(errors) > 0 {
		fmt.Printf("Warning: %d conversations failed to process\n", len(errors))
		// Print first few errors as examples
		for i, err := range errors {
			if i >= 5 { // Limit to first 5 errors to avoid spam
				fmt.Printf("... and %d more errors\n", len(errors)-5)
				break
			}
			fmt.Printf("  - %v\n", err)
		}
	}

	return nil
}

// conversationWorker processes conversation jobs from the job channel
func (p *ChatGPTParser) conversationWorker(wg *sync.WaitGroup, jobChan <-chan conversationJob, resultChan chan<- error, progressChan chan<- int, callback func(models.ChatGPTConversation) error) {
	defer wg.Done()

	for job := range jobChan {
		var err error

		// Convert raw conversation to standard format
		conv, convErr := p.convertRawConversation(job.rawConv)
		if convErr != nil {
			err = fmt.Errorf("failed to convert conversation %d: %w", job.index, convErr)
		} else {
			// Warn about empty mappings but don't fail
			if len(conv.Mapping) == 0 {
				fmt.Printf("Warning: conversation %s has empty mapping after conversion\n", conv.ID)
			}

			// Call the callback function
			if callbackErr := callback(conv); callbackErr != nil {
				err = fmt.Errorf("callback failed for conversation %s: %w", conv.ID, callbackErr)
			}
		}

		resultChan <- err
		progressChan <- 1 // Signal one conversation processed
	}
}

// progressReporter reports progress of conversation processing
func (p *ChatGPTParser) progressReporter(wg *sync.WaitGroup, progressChan <-chan int, total int) {
	defer wg.Done()

	processed := 0
	for range progressChan {
		processed++
		if processed%100 == 0 || processed == total {
			fmt.Printf("Processed %d/%d conversations...\n", processed, total)
		}
	}
}

// convertRawConversation converts the raw ChatGPT format to our standard format
func (p *ChatGPTParser) convertRawConversation(raw models.ChatGPTConversationRaw) (models.ChatGPTConversation, error) {
	// Use GUID as fallback title if title is empty
	title := raw.Title
	if title == "" {
		title = fmt.Sprintf("Conversation-%s", raw.ID[:8]) // Use first 8 chars of GUID
	}
	
	conv := models.ChatGPTConversation{
		ID:             raw.ID,
		Title:          title,
		CreateTime:     raw.CreateTime,
		UpdateTime:     raw.UpdateTime,
		Mapping:        make(map[string]models.ChatGPTNode),
		CurrentNode:    raw.CurrentNode,
		ConversationID: raw.ConversationID,
	}

	// Convert mapping with proper error handling
	for nodeID, rawNode := range raw.Mapping {
		node := models.ChatGPTNode{
			ID:       rawNode.ID,
			Parent:   rawNode.Parent,
			Children: rawNode.Children,
		}

		// Handle message conversion with type safety
		if rawNode.Message != nil {
			message, err := p.convertRawMessage(*rawNode.Message)
			if err != nil {
				// Log but don't fail - add the node without the message
				// This preserves the tree structure for navigation
				fmt.Printf("Warning: failed to convert message in node %s: %v\n", nodeID, err)
				// Don't continue here - we still want to add the node to preserve tree structure
				// The node.Message will remain nil
			} else {
				node.Message = &message
			}
		}

		conv.Mapping[nodeID] = node
	}

	return conv, nil
}

// convertRawMessage converts raw message format with flexible content handling
func (p *ChatGPTParser) convertRawMessage(raw models.ChatGPTMessageRaw) (models.ChatGPTMessage, error) {
	message := models.ChatGPTMessage{
		ID:         raw.ID,
		Author:     raw.Author,
		CreateTime: raw.CreateTime,
		UpdateTime: raw.UpdateTime,
		Status:     raw.Status,
		Metadata:   raw.Metadata,
	}

	// Handle flexible content format
	content := models.ChatGPTContent{
		ContentType: raw.Content.ContentType,
	}

	// Convert parts based on their actual type
	switch v := raw.Content.Parts.(type) {
	case []interface{}:
		// Handle array of mixed types
		for _, part := range v {
			switch partVal := part.(type) {
			case string:
				content.Parts = append(content.Parts, partVal)
			case map[string]interface{}:
				// Handle object content (images, etc.)
				if str, ok := partVal["text"].(string); ok {
					content.Parts = append(content.Parts, str)
				} else {
					// Convert object to string representation
					partStr := fmt.Sprintf("[Object: %v]", partVal)
					content.Parts = append(content.Parts, partStr)
				}
			default:
				// Convert other types to string
				content.Parts = append(content.Parts, fmt.Sprintf("%v", partVal))
			}
		}
	case []string:
		// Handle string array
		content.Parts = v
	case string:
		// Handle single string
		content.Parts = []string{v}
	default:
		// Handle unexpected format
		content.Parts = []string{fmt.Sprintf("%v", v)}
	}

	message.Content = content
	return message, nil
}

// ParseUserInfo parses user.json file
func (p *ChatGPTParser) ParseUserInfo() (*models.ChatGPTUser, error) {
	filePath := filepath.Join(p.inputPath, "chat-gpt-2025-06-13", "user.json")
	
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open user.json: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read user.json: %w", err)
	}

	var user models.ChatGPTUser
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user.json: %w", err)
	}

	return &user, nil
}

// GetMediaFiles scans for media files in the ChatGPT export
func (p *ChatGPTParser) GetMediaFiles() (*models.ChatGPTMediaInfo, error) {
	baseDir := filepath.Join(p.inputPath, "chat-gpt-2025-06-13")
	
	mediaInfo := &models.ChatGPTMediaInfo{
		Images:           []models.MediaFile{},
		DalleGenerations: []models.MediaFile{},
		UserUploads:      []models.MediaFile{},
		AudioConversations: []models.AudioConversation{},
	}

	// Scan main directory for images
	err := p.scanDirectoryForImages(baseDir, &mediaInfo.Images)
	if err != nil {
		return nil, fmt.Errorf("failed to scan main directory: %w", err)
	}

	// Scan dalle-generations
	dalleDir := filepath.Join(baseDir, "dalle-generations")
	if _, err := os.Stat(dalleDir); err == nil {
		err = p.scanDirectoryForImages(dalleDir, &mediaInfo.DalleGenerations)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dalle-generations: %w", err)
		}
	}

	// Scan user uploads
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read base directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "user-") {
			userDir := filepath.Join(baseDir, entry.Name())
			err = p.scanDirectoryForImages(userDir, &mediaInfo.UserUploads)
			if err != nil {
				fmt.Printf("Warning: failed to scan user directory %s: %v\n", entry.Name(), err)
			}
		}
	}

	// Scan for audio conversations
	for _, entry := range entries {
		if entry.IsDir() && len(entry.Name()) > 20 && !strings.HasPrefix(entry.Name(), "user-") && !strings.HasPrefix(entry.Name(), "dalle-") {
			// Likely a conversation ID directory
			audioDir := filepath.Join(baseDir, entry.Name(), "audio")
			if _, err := os.Stat(audioDir); err == nil {
				audioConv, err := p.scanAudioDirectory(entry.Name(), audioDir)
				if err != nil {
					fmt.Printf("Warning: failed to scan audio directory %s: %v\n", entry.Name(), err)
				} else {
					mediaInfo.AudioConversations = append(mediaInfo.AudioConversations, *audioConv)
				}
			}
		}
	}

	return mediaInfo, nil
}

// scanDirectoryForImages scans a directory for image files
func (p *ChatGPTParser) scanDirectoryForImages(dir string, images *[]models.MediaFile) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			*images = append(*images, models.MediaFile{
				Name:     name,
				Path:     filepath.Join(dir, name),
				Size:     info.Size(),
				Modified: info.ModTime(),
			})
		}
	}

	return nil
}

// scanAudioDirectory scans for audio files in a conversation directory
func (p *ChatGPTParser) scanAudioDirectory(conversationID, audioDir string) (*models.AudioConversation, error) {
	entries, err := os.ReadDir(audioDir)
	if err != nil {
		return nil, err
	}

	audioConv := &models.AudioConversation{
		ConversationID: conversationID,
		AudioFiles:     []models.MediaFile{},
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".wav" || ext == ".mp3" || ext == ".m4a" {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			audioConv.AudioFiles = append(audioConv.AudioFiles, models.MediaFile{
				Name:     name,
				Path:     filepath.Join(audioDir, name),
				Size:     info.Size(),
				Modified: info.ModTime(),
			})
		}
	}

	return audioConv, nil
}