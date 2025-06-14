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
	ChatMessages   []ClaudeMessage        `json:"chat_messages"`
	Settings       map[string]interface{} `json:"settings,omitempty"`
}

// ClaudeMessage represents a single message in Claude format
type ClaudeMessage struct {
	UUID      string                 `json:"uuid"`
	Text      string                 `json:"text"`
	Sender    string                 `json:"sender"`
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
	UUID         string           `json:"uuid"`
	Name         string           `json:"name"`
	Description  string           `json:"description,omitempty"`
	CreatedAt    string           `json:"created_at"`
	UpdatedAt    string           `json:"updated_at"`
	ArchivedAt   string           `json:"archived_at,omitempty"`
	Docs         []ClaudeDocument `json:"docs,omitempty"`
}

// ClaudeDocument represents a project document
type ClaudeDocument struct {
	UUID      string `json:"uuid"`
	Filename  string `json:"filename"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
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

// ChatGPTConversationRaw represents the raw ChatGPT conversation format
type ChatGPTConversationRaw struct {
	ID              string                     `json:"id"`
	Title           string                     `json:"title"`
	CreateTime      float64                    `json:"create_time"`
	UpdateTime      float64                    `json:"update_time"`
	Mapping         map[string]ChatGPTNodeRaw  `json:"mapping"`
	CurrentNode     string                     `json:"current_node"`
	ConversationID  string                     `json:"conversation_id"`
}

// ChatGPTNodeRaw represents a raw node in the ChatGPT conversation tree
type ChatGPTNodeRaw struct {
	ID       string                `json:"id"`
	Message  *ChatGPTMessageRaw    `json:"message,omitempty"`
	Parent   string                `json:"parent,omitempty"`
	Children []string              `json:"children,omitempty"`
}

// ChatGPTMessageRaw represents a raw message with flexible content parsing
type ChatGPTMessageRaw struct {
	ID         string                 `json:"id"`
	Author     ChatGPTAuthor          `json:"author"`
	CreateTime float64                `json:"create_time"`
	UpdateTime float64                `json:"update_time"`
	Content    ChatGPTContentRaw      `json:"content"`
	Status     string                 `json:"status"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ChatGPTContentRaw represents raw content with flexible parts handling
type ChatGPTContentRaw struct {
	ContentType string      `json:"content_type"`
	Parts       interface{} `json:"parts"` // Can be []string, []interface{}, or string
}

// ChatGPTUser represents user information
type ChatGPTUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Image    string `json:"image"`
	Picture  string `json:"picture"`
	Idp      string `json:"idp"`
	Iat      int64  `json:"iat"`
	Mfa      bool   `json:"mfa"`
	Groups   []string `json:"groups"`
}

// MediaFile represents a media file
type MediaFile struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

// AudioConversation represents an audio conversation
type AudioConversation struct {
	ConversationID string      `json:"conversation_id"`
	AudioFiles     []MediaFile `json:"audio_files"`
}

// ChatGPTMediaInfo represents all media files in a ChatGPT export
type ChatGPTMediaInfo struct {
	Images             []MediaFile         `json:"images"`
	DalleGenerations   []MediaFile         `json:"dalle_generations"`
	UserUploads        []MediaFile         `json:"user_uploads"`
	AudioConversations []AudioConversation `json:"audio_conversations"`
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