package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"chat-transformer/internal/models"
)

// Parser handles parsing of large JSON files
type Parser struct {
	inputPath string
}

// New creates a new parser instance
func New(inputPath string) *Parser {
	return &Parser{
		inputPath: inputPath,
	}
}

// ParseClaudeConversations parses Claude conversations.json file
func (p *Parser) ParseClaudeConversations(callback func(models.ClaudeConversation) error) error {
	file, err := os.Open(p.inputPath + "/claude-2025-06-13/conversations.json")
	if err != nil {
		return fmt.Errorf("failed to open Claude conversations file: %w", err)
	}
	defer file.Close()

	// Read the entire file into memory
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read Claude conversations file: %w", err)
	}

	// Parse JSON array directly
	var conversations []models.ClaudeConversation
	if err := json.Unmarshal(data, &conversations); err != nil {
		return fmt.Errorf("failed to parse Claude conversations JSON: %w", err)
	}

	// Process each conversation
	for _, conv := range conversations {
		if err := callback(conv); err != nil {
			fmt.Printf("Warning: callback failed for Claude conversation %s: %v\n", conv.UUID, err)
		}
	}

	return nil
}

// ParseClaudeProjects parses Claude projects.json file
func (p *Parser) ParseClaudeProjects() ([]models.ClaudeProject, error) {
	file, err := os.Open(p.inputPath + "/claude-2025-06-13/projects.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open Claude projects file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read Claude projects file: %w", err)
	}

	var projects []models.ClaudeProject
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse Claude projects: %w", err)
	}

	return projects, nil
}

// ParseChatGPTConversations parses ChatGPT conversations.json file
func (p *Parser) ParseChatGPTConversations(callback func(models.ChatGPTConversation) error) error {
	file, err := os.Open(p.inputPath + "/chat-gpt-2025-06-13/conversations.json")
	if err != nil {
		return fmt.Errorf("failed to open ChatGPT conversations file: %w", err)
	}
	defer file.Close()

	// Read the entire file into memory
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read ChatGPT conversations file: %w", err)
	}

	// Parse JSON array directly
	var conversations []models.ChatGPTConversation
	if err := json.Unmarshal(data, &conversations); err != nil {
		return fmt.Errorf("failed to parse ChatGPT conversations JSON: %w", err)
	}

	// Process each conversation
	for _, conv := range conversations {
		if err := callback(conv); err != nil {
			fmt.Printf("Warning: callback failed for ChatGPT conversation %s: %v\n", conv.ID, err)
		}
	}

	return nil
}

// ConvertClaudeToStandard converts Claude conversation to standard format
func ConvertClaudeToStandard(claude models.ClaudeConversation, projects map[string]models.ClaudeProject) models.Conversation {
	createdAt, _ := time.Parse(time.RFC3339, claude.CreatedAt)
	updatedAt, _ := time.Parse(time.RFC3339, claude.UpdatedAt)

	// Determine project name
	projectName := ""
	if claude.ProjectUUID != "" {
		if project, exists := projects[claude.ProjectUUID]; exists {
			projectName = project.Name
		}
	}

	// Convert messages
	var messages []models.Message
	participants := make(map[string]bool)
	hasCode := false
	hasMedia := false

	for _, msg := range claude.Messages {
		msgTime, _ := time.Parse(time.RFC3339, msg.CreatedAt)
		
		// Extract text content
		var content strings.Builder
		for _, c := range msg.Content {
			if c.Type == "text" {
				content.WriteString(c.Text)
			} else if c.Type == "image" {
				hasMedia = true
				content.WriteString(fmt.Sprintf("[Image: %s]", c.URL))
			}
		}

		contentText := content.String()
		if strings.Contains(contentText, "```") || strings.Contains(contentText, "`") {
			hasCode = true
		}

		author := msg.Role
		if author == "assistant" {
			author = "Claude"
		} else if author == "user" {
			author = "User"
		}
		participants[author] = true

		messages = append(messages, models.Message{
			ID:        msg.UUID,
			Author:    author,
			Content:   contentText,
			Timestamp: msgTime,
			Metadata:  msg.Metadata,
		})
	}

	// Convert participants map to slice
	var partList []string
	for p := range participants {
		partList = append(partList, p)
	}

	metadata := models.ConversationMetadata{
		ID:           claude.UUID,
		Title:        claude.Name,
		Platform:     "claude",
		Project:      projectName,
		CreatedDate:  createdAt,
		LastModified: updatedAt,
		MessageCount: len(messages),
		Participants: partList,
		Topics:       extractTopics(claude.Name),
		HasCode:      hasCode,
		HasMedia:     hasMedia,
	}

	return models.Conversation{
		Metadata: metadata,
		Messages: messages,
	}
}

// ConvertChatGPTToStandard converts ChatGPT conversation to standard format
func ConvertChatGPTToStandard(chatgpt models.ChatGPTConversation) models.Conversation {
	createdAt := time.Unix(int64(chatgpt.CreateTime), 0)
	updatedAt := time.Unix(int64(chatgpt.UpdateTime), 0)

	// Extract messages from the conversation tree
	var messages []models.Message
	participants := make(map[string]bool)
	hasCode := false
	hasMedia := false

	// Build message chain from the tree structure
	visitedNodes := make(map[string]bool)
	var extractMessages func(nodeID string)
	
	extractMessages = func(nodeID string) {
		if nodeID == "" || visitedNodes[nodeID] {
			return
		}
		
		visitedNodes[nodeID] = true
		node, exists := chatgpt.Mapping[nodeID]
		if !exists || node.Message == nil {
			return
		}

		msg := node.Message
		msgTime := time.Unix(int64(msg.CreateTime), 0)
		
		// Extract content
		var content strings.Builder
		for _, part := range msg.Content.Parts {
			content.WriteString(part)
			content.WriteString(" ")
		}
		
		contentText := strings.TrimSpace(content.String())
		if contentText == "" {
			// Process children
			for _, childID := range node.Children {
				extractMessages(childID)
			}
			return
		}

		if strings.Contains(contentText, "```") || strings.Contains(contentText, "`") {
			hasCode = true
		}

		author := msg.Author.Role
		if author == "assistant" {
			author = "ChatGPT"
		} else if author == "user" {
			author = "User"
		}
		participants[author] = true

		messages = append(messages, models.Message{
			ID:        msg.ID,
			Author:    author,
			Content:   contentText,
			Timestamp: msgTime,
			Metadata:  msg.Metadata,
		})

		// Process children
		for _, childID := range node.Children {
			extractMessages(childID)
		}
	}

	// Start from root nodes (nodes with no parent)
	for nodeID, node := range chatgpt.Mapping {
		if node.Parent == "" {
			extractMessages(nodeID)
		}
	}

	// Sort messages by timestamp
	for i := 0; i < len(messages)-1; i++ {
		for j := i + 1; j < len(messages); j++ {
			if messages[i].Timestamp.After(messages[j].Timestamp) {
				messages[i], messages[j] = messages[j], messages[i]
			}
		}
	}

	// Convert participants map to slice
	var partList []string
	for p := range participants {
		partList = append(partList, p)
	}

	metadata := models.ConversationMetadata{
		ID:           chatgpt.ID,
		Title:        chatgpt.Title,
		Platform:     "chatgpt",
		CreatedDate:  createdAt,
		LastModified: updatedAt,
		MessageCount: len(messages),
		Participants: partList,
		Topics:       extractTopics(chatgpt.Title),
		HasCode:      hasCode,
		HasMedia:     hasMedia,
	}

	return models.Conversation{
		Metadata: metadata,
		Messages: messages,
	}
}

// extractTopics extracts basic topics from conversation title
func extractTopics(title string) []string {
	// Simple topic extraction - can be enhanced with NLP
	topics := []string{}
	
	// Common programming topics
	programmingKeywords := []string{
		"python", "javascript", "go", "golang", "react", "node", "api",
		"database", "sql", "web", "frontend", "backend", "code", "programming",
		"debug", "error", "function", "class", "algorithm", "data", "structure",
	}
	
	titleLower := strings.ToLower(title)
	for _, keyword := range programmingKeywords {
		if strings.Contains(titleLower, keyword) {
			topics = append(topics, keyword)
		}
	}
	
	// If no specific topics found, categorize generically
	if len(topics) == 0 {
		if strings.Contains(titleLower, "help") || strings.Contains(titleLower, "question") {
			topics = append(topics, "help")
		} else {
			topics = append(topics, "general")
		}
	}
	
	return topics
}