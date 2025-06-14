package processor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"chat-transformer/internal/models"
)

// copyChatGPTMediaFiles copies media files to organized folders when copyMedia flag is set
func (p *Processor) copyChatGPTMediaFiles(mediaInfo *models.ChatGPTMediaInfo) error {
	mediaBase := filepath.Join(p.outputPath, "chatgpt", "media")

	// Create organized subdirectories
	dirs := []string{
		"images",
		"dalle-generations",
		"user-uploads",
		"audio-conversations",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(mediaBase, dir), 0755); err != nil {
			return fmt.Errorf("failed to create media directory %s: %w", dir, err)
		}
	}

	// Copy images
	for _, file := range mediaInfo.Images {
		destPath := filepath.Join(mediaBase, "images", file.Name)
		if err := p.copyFile(file.Path, destPath); err != nil {
			fmt.Printf("Warning: failed to copy image %s: %v\n", file.Name, err)
		}
	}

	// Copy DALL-E generations
	for _, file := range mediaInfo.DalleGenerations {
		destPath := filepath.Join(mediaBase, "dalle-generations", file.Name)
		if err := p.copyFile(file.Path, destPath); err != nil {
			fmt.Printf("Warning: failed to copy DALL-E image %s: %v\n", file.Name, err)
		}
	}

	// Copy user uploads
	for _, file := range mediaInfo.UserUploads {
		destPath := filepath.Join(mediaBase, "user-uploads", file.Name)
		if err := p.copyFile(file.Path, destPath); err != nil {
			fmt.Printf("Warning: failed to copy user upload %s: %v\n", file.Name, err)
		}
	}

	// Copy audio conversations
	for _, audioConv := range mediaInfo.AudioConversations {
		convDir := filepath.Join(mediaBase, "audio-conversations", audioConv.ConversationID)
		if err := os.MkdirAll(convDir, 0755); err != nil {
			fmt.Printf("Warning: failed to create audio conversation directory %s: %v\n", audioConv.ConversationID, err)
			continue
		}

		for _, file := range audioConv.AudioFiles {
			destPath := filepath.Join(convDir, file.Name)
			if err := p.copyFile(file.Path, destPath); err != nil {
				fmt.Printf("Warning: failed to copy audio file %s: %v\n", file.Name, err)
			}
		}
	}

	// Create helpful README files for media processing
	return p.createMediaREADMEs(mediaBase)
}

// copyFile copies a file from src to dst
func (p *Processor) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// createMediaREADMEs creates helpful README files for media processing
func (p *Processor) createMediaREADMEs(mediaBase string) error {
	readmes := map[string]string{
		"images/README.md": `# Images from ChatGPT Conversations

This directory contains images that were uploaded to or referenced in ChatGPT conversations.

## Processing Suggestions

### Image Description and OCR
` + "```bash" + `
# Use Claude, ChatGPT, or local tools to describe images
python -c "
import base64
from pathlib import Path

def encode_image(image_path):
    with open(image_path, 'rb') as f:
        return base64.b64encode(f.read()).decode('utf-8')

# For each image file
for img in Path('.').glob('*.{jpg,jpeg,png,webp}'):
    encoded = encode_image(img)
    print(f'Image: {img.name}')
    # Send to vision model for description
"
` + "```" + `

### Batch Processing with Vision APIs
` + "```bash" + `
# Using OpenAI Vision API
for img in *.{jpg,jpeg,png,webp}; do
    echo "Processing $img..."
    curl -X POST https://api.openai.com/v1/chat/completions \
      -H "Authorization: Bearer $OPENAI_API_KEY" \
      -H "Content-Type: application/json" \
      -d '{
        "model": "gpt-4-vision-preview",
        "messages": [{"role": "user", "content": [
          {"type": "text", "text": "Describe this image in detail"},
          {"type": "image_url", "image_url": {"url": "data:image/jpeg;base64,'$(base64 -i "$img")'"}}
        ]}]
      }' > "${img%.jpg}_description.json"
done
` + "```" + `
`,

		"dalle-generations/README.md": `# DALL-E Generated Images

This directory contains AI-generated images from DALL-E in ChatGPT conversations.

## Processing Suggestions

### Metadata Extraction
DALL-E images often contain generation prompts in their metadata or filenames.

` + "```bash" + `
# Extract metadata from images
exiftool *.{jpg,jpeg,png,webp} > metadata.txt

# Look for generation prompts in conversation history
grep -r "dall-e\|generate.*image\|create.*image" ../../chats/
` + "```" + `

### Prompt Recreation
` + "```bash" + `
# Create a catalog of prompts and results
python -c "
import json
from pathlib import Path

# Load conversations and find DALL-E references
for conv_file in Path('../../chats').rglob('*.json'):
    with open(conv_file) as f:
        data = json.load(f)
        for msg in data.get('messages', []):
            if 'dall' in msg.get('content', '').lower():
                print(f'Conversation: {conv_file.name}')
                print(f'Prompt: {msg[\"content\"][:100]}...')
                print('---')
"
` + "```" + `
`,

		"user-uploads/README.md": `# User Uploaded Files

This directory contains files uploaded by users to ChatGPT conversations.

## Processing Suggestions

### Document Processing
` + "```bash" + `
# Extract text from PDFs
for pdf in *.pdf; do
    pdftotext "$pdf" "${pdf%.pdf}.txt"
done

# Process images with OCR
for img in *.{jpg,jpeg,png}; do
    tesseract "$img" "${img%.*}_ocr" -l eng
done
` + "```" + `

### File Analysis
` + "```bash" + `
# Get file types and sizes
file * > file_types.txt
du -h * > file_sizes.txt

# For spreadsheets
for xlsx in *.xlsx; do
    # Convert to CSV for analysis
    ssconvert "$xlsx" "${xlsx%.xlsx}.csv"
done
` + "```" + `
`,

		"audio-conversations/README.md": `# Audio Conversations

This directory contains audio files from ChatGPT voice conversations.

## Processing Suggestions

### Transcription with Whisper
` + "```bash" + `
# Install OpenAI Whisper
pip install openai-whisper

# Transcribe all audio files
for audio_dir in */; do
    echo "Processing conversation: $audio_dir"
    cd "$audio_dir"
    for audio in *.wav; do
        whisper "$audio" --output_format txt --output_dir transcripts/
    done
    cd ..
done
` + "```" + `

### Batch Processing with APIs
` + "```bash" + `
# Using OpenAI Whisper API
for audio_dir in */; do
    cd "$audio_dir"
    for audio in *.wav; do
        curl -X POST https://api.openai.com/v1/audio/transcriptions \
          -H "Authorization: Bearer $OPENAI_API_KEY" \
          -H "Content-Type: multipart/form-data" \
          -F file="@$audio" \
          -F model="whisper-1" \
          > "${audio%.wav}_transcript.json"
    done
    cd ..
done
` + "```" + `

### Analysis Scripts
` + "```python" + `
# Python script to organize transcripts
import json
import os
from pathlib import Path

def process_conversation_audio(conv_dir):
    transcripts = []
    for audio_file in Path(conv_dir).glob("*_transcript.json"):
        with open(audio_file) as f:
            data = json.load(f)
            transcripts.append({
                "file": audio_file.stem,
                "text": data.get("text", ""),
                "timestamp": audio_file.stat().st_mtime
            })
    
    # Sort by timestamp and combine
    transcripts.sort(key=lambda x: x["timestamp"])
    full_conversation = "\n\n".join([t["text"] for t in transcripts])
    
    # Save combined transcript
    with open(f"{conv_dir}/full_conversation.txt", "w") as f:
        f.write(full_conversation)

# Process all conversation directories
for conv_dir in Path(".").glob("*/"):
    if conv_dir.is_dir():
        process_conversation_audio(conv_dir)
` + "```" + `
`,
	}

	for path, content := range readmes {
		fullPath := filepath.Join(mediaBase, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", path, err)
		}
	}

	return nil
}