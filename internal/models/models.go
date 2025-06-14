package models

import (
	"time"
)

// ConversationMetadata represents metadata for any conversation
type ConversationMetadata struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Platform     string    `json:"platform"`
	Project      string    `json:"project,omitempty"`
	CreatedDate  time.Time `json:"created_date"`
	LastModified time.Time `json:"last_modified"`
	MessageCount int       `json:"message_count"`
	Participants []string  `json:"participants"`
	Topics       []string  `json:"topics"`
	HasCode      bool      `json:"has_code"`
	HasMedia     bool      `json:"has_media"`
	FilePath     string    `json:"file_path"`
}

// Message represents a single message in a conversation
type Message struct {
	ID        string                 `json:"id"`
	Author    string                 `json:"author"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Conversation represents a full conversation
type Conversation struct {
	Metadata ConversationMetadata `json:"metadata"`
	Messages []Message            `json:"messages"`
}

// ClaudeConversation represents the structure of Claude conversations
type ClaudeConversation struct {
	UUID           string                 `json:"uuid"`
	Name           string                 `json:"name"`
	CreatedAt      string                 `json:"created_at"`
	UpdatedAt      string                 `json:"updated_at"`
	ProjectUUID    string                 `json:"project_uuid,omitempty"`
	Messages       []ClaudeMessage        `json:"messages"`
	Settings       map[string]interface{} `json:"settings,omitempty"`
}

// ClaudeMessage represents a single message in Claude format
type ClaudeMessage struct {
	UUID      string                 `json:"uuid"`
	Role      string                 `json:"role"`
	Content   []ClaudeContent        `json:"content"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ClaudeContent represents content within a Claude message
type ClaudeContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	URL  string `json:"url,omitempty"`
}

// ClaudeProject represents a project from projects.json
type ClaudeProject struct {
	UUID         string    `json:"uuid"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	CreatedAt    string    `json:"created_at"`
	UpdatedAt    string    `json:"updated_at"`
	ArchivedAt   string    `json:"archived_at,omitempty"`
}

// ChatGPTConversation represents the structure of ChatGPT conversations
type ChatGPTConversation struct {
	ID              string                  `json:"id"`
	Title           string                  `json:"title"`
	CreateTime      float64                 `json:"create_time"`
	UpdateTime      float64                 `json:"update_time"`
	Mapping         map[string]ChatGPTNode  `json:"mapping"`
	CurrentNode     string                  `json:"current_node"`
	ConversationID  string                  `json:"conversation_id"`
}

// ChatGPTNode represents a node in the ChatGPT conversation tree
type ChatGPTNode struct {
	ID       string            `json:"id"`
	Message  *ChatGPTMessage   `json:"message,omitempty"`
	Parent   string            `json:"parent,omitempty"`
	Children []string          `json:"children,omitempty"`
}

// ChatGPTMessage represents a single message in ChatGPT format
type ChatGPTMessage struct {
	ID         string                 `json:"id"`
	Author     ChatGPTAuthor          `json:"author"`
	CreateTime float64                `json:"create_time"`
	UpdateTime float64                `json:"update_time"`
	Content    ChatGPTContent         `json:"content"`
	Status     string                 `json:"status"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ChatGPTAuthor represents the author of a ChatGPT message
type ChatGPTAuthor struct {
	Role     string                 `json:"role"`
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChatGPTContent represents the content of a ChatGPT message
type ChatGPTContent struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

// Index represents various index structures
type Index struct {
	Conversations []ConversationMetadata    `json:"conversations"`
	LastUpdated   time.Time                 `json:"last_updated"`
}

// TopicIndex represents topic-based indexing
type TopicIndex struct {
	Topics      map[string][]string `json:"topics"` // topic -> conversation IDs
	LastUpdated time.Time           `json:"last_updated"`
}

// MediaIndex represents media file indexing
type MediaIndex struct {
	Media       []MediaItem `json:"media"`
	LastUpdated time.Time   `json:"last_updated"`
}

// MediaItem represents a media file reference
type MediaItem struct {
	ID             string    `json:"id"`
	Type           string    `json:"type"` // audio, image
	OriginalPath   string    `json:"original_path"`
	NewPath        string    `json:"new_path"`
	ConversationID string    `json:"conversation_id"`
	MessageID      string    `json:"message_id,omitempty"`
	Prompt         string    `json:"prompt,omitempty"` // for DALL-E images
	CreatedAt      time.Time `json:"created_at"`
}