# Chat Export Transformer

A Go application that transforms raw ChatGPT and Claude chat exports into a structured, searchable, and digestible format for analysis and reference.

## Features

- **Stream Processing**: Handles large JSON files (100MB+) without loading entirely into memory
- **Structured Organization**: Organizes conversations by date, project, and topic
- **Cross-Platform Support**: Processes both Claude and ChatGPT exports
- **Search Indexes**: Generates comprehensive indexes for discovery and search
- **Media Handling**: Organizes and links media files to conversations
- **Metadata Extraction**: Enriches conversations with topics, code detection, and more

## Installation

```bash
git clone <repository>
cd chat-transformer
go mod tidy
go build -o chat-transformer
```

## Usage

### Basic Usage (assumes you're in the @wisdom folder)
```bash
./chat-transformer
```

### Custom Input/Output Paths
```bash
./chat-transformer -i /path/to/raw/exports -o /path/to/output

# Or with long flags
./chat-transformer --input-folder /path/to/raw/exports --output-folder /path/to/output
```

## Input Structure

The application expects the following input structure:

```
raw/
├── claude-2025-06-13/
│   ├── conversations.json (large file with all conversations)
│   ├── projects.json (project metadata)
│   └── users.json (user information)
└── chat-gpt-2025-06-13/
    ├── conversations.json (large file with all conversations)
    ├── chat.html (HTML export)
    └── [media files and directories]
```

## Output Structure

The application creates the following organized structure:

```
expanded/
├── claude/
│   ├── projects/
│   │   ├── [project-name]/
│   │   │   ├── metadata.json
│   │   │   └── conversations/
│   │   │       ├── YYYY-MM-DD_conversation-title.json
│   │   │       └── ...
│   │   └── ...
│   ├── general-chats/
│   │   ├── YYYY/
│   │   │   ├── MM/
│   │   │   │   ├── conversation-title_YYYY-MM-DD.json
│   │   │   │   └── ...
│   │   │   └── ...
│   │   └── ...
│   └── index/
│       ├── conversations_index.json
│       └── topics_index.json
├── chatgpt/
│   ├── conversations/
│   │   ├── YYYY/
│   │   │   └── MM/
│   │   │       ├── conversation-title_YYYY-MM-DD/
│   │   │       │   ├── conversation.json
│   │   │       │   ├── audio/ (if any)
│   │   │       │   └── images/ (if any)
│   │   │       └── ...
│   │   └── ...
│   └── index/
│       ├── conversations_index.json
│       └── media_index.json
└── unified/
    ├── conversations_index.json (all conversations)
    ├── topics_index.json (cross-platform topics)
    └── timeline.json (chronological view)
```

## Data Models

### Conversation Metadata
```json
{
  "id": "conversation-uuid",
  "title": "extracted or generated title",
  "platform": "claude|chatgpt",
  "project": "project-name (if applicable)",
  "created_date": "2024-01-01T00:00:00Z",
  "last_modified": "2024-01-01T00:00:00Z",
  "message_count": 42,
  "participants": ["User", "Claude"],
  "topics": ["programming", "web development"],
  "has_code": true,
  "has_media": false,
  "file_path": "relative/path/to/conversation.json"
}
```

### Message Structure
```json
{
  "id": "message-uuid",
  "author": "User|Claude|ChatGPT",
  "content": "message content",
  "timestamp": "2024-01-01T00:00:00Z",
  "metadata": {}
}
```

## Index Files

### Conversation Index
- Lists all conversations with metadata
- Separate indexes for Claude, ChatGPT, and unified
- Enables quick filtering and search

### Topic Index
- Groups conversations by automatically extracted topics
- Cross-platform topic discovery
- Enables thematic browsing

### Timeline Index
- Chronological ordering of all conversations
- Date range information
- Historical analysis support

## Performance

- **Memory Efficient**: Streams large JSON files
- **Fast Processing**: Parallel processing where possible
- **Scalable**: Handles hundreds of MB of chat data
- **Resumable**: Checkpoint capability for large transformations

## Future Enhancements

- Media file processing and organization
- Enhanced topic extraction using NLP
- Full-text search index generation
- MCP (Model Context Protocol) integration
- Web interface for browsing conversations

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details.