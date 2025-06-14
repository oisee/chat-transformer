package indexer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"chat-transformer/internal/models"
)

// Indexer handles creation of search and discovery indexes
type Indexer struct {
	outputPath    string
	conversations []models.ConversationMetadata
	topics        map[string][]string // topic -> conversation IDs
	mutex         sync.RWMutex        // protects conversations and topics maps
}

// New creates a new indexer instance
func New(outputPath string) *Indexer {
	return &Indexer{
		outputPath:    outputPath,
		conversations: make([]models.ConversationMetadata, 0),
		topics:        make(map[string][]string),
	}
}

// AddConversation adds a conversation to the index
func (idx *Indexer) AddConversation(metadata models.ConversationMetadata) {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()
	
	idx.conversations = append(idx.conversations, metadata)
	
	// Add to topic index
	for _, topic := range metadata.Topics {
		if _, exists := idx.topics[topic]; !exists {
			idx.topics[topic] = make([]string, 0)
		}
		idx.topics[topic] = append(idx.topics[topic], metadata.ID)
	}
}

// GenerateIndexes generates all index files
func (idx *Indexer) GenerateIndexes() error {
	// Generate main conversation index
	if err := idx.generateConversationIndex(); err != nil {
		return err
	}

	// Generate topic index
	if err := idx.generateTopicIndex(); err != nil {
		return err
	}

	// Generate unified timeline
	if err := idx.generateTimeline(); err != nil {
		return err
	}

	return nil
}

// generateConversationIndex creates the main conversation index
func (idx *Indexer) generateConversationIndex() error {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()
	
	// Claude index
	claudeConvs := make([]models.ConversationMetadata, 0)
	chatgptConvs := make([]models.ConversationMetadata, 0)

	for _, conv := range idx.conversations {
		if conv.Platform == "claude" {
			claudeConvs = append(claudeConvs, conv)
		} else if conv.Platform == "chatgpt" {
			chatgptConvs = append(chatgptConvs, conv)
		}
	}

	// Save Claude index
	claudeIndex := models.Index{
		Conversations: claudeConvs,
		LastUpdated:   time.Now(),
	}
	if err := idx.saveIndex(claudeIndex, "claude/index/conversations_index.json"); err != nil {
		return err
	}

	// Save ChatGPT index
	chatgptIndex := models.Index{
		Conversations: chatgptConvs,
		LastUpdated:   time.Now(),
	}
	if err := idx.saveIndex(chatgptIndex, "chatgpt/index/conversations_index.json"); err != nil {
		return err
	}

	// Save unified index
	unifiedIndex := models.Index{
		Conversations: idx.conversations,
		LastUpdated:   time.Now(),
	}
	return idx.saveIndex(unifiedIndex, "unified/conversations_index.json")
}

// generateTopicIndex creates topic-based indexes
func (idx *Indexer) generateTopicIndex() error {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()
	
	topicIndex := models.TopicIndex{
		Topics:      idx.topics,
		LastUpdated: time.Now(),
	}

	return idx.saveIndex(topicIndex, "unified/topics_index.json")
}

// generateTimeline creates a chronological timeline
func (idx *Indexer) generateTimeline() error {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()
	
	// Sort conversations by date
	sorted := make([]models.ConversationMetadata, len(idx.conversations))
	copy(sorted, idx.conversations)

	// Simple bubble sort by creation date
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].CreatedDate.After(sorted[j].CreatedDate) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	timeline := map[string]interface{}{
		"conversations": sorted,
		"total_count":   len(sorted),
		"date_range": map[string]interface{}{
			"earliest": sorted[0].CreatedDate,
			"latest":   sorted[len(sorted)-1].CreatedDate,
		},
		"last_updated": time.Now(),
	}

	return idx.saveIndex(timeline, "unified/timeline.json")
}

// saveIndex saves an index to disk
func (idx *Indexer) saveIndex(data interface{}, relativePath string) error {
	fullPath := filepath.Join(idx.outputPath, relativePath)
	
	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}