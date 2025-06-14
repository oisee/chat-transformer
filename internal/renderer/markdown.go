package renderer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"chat-transformer/internal/models"
)

const (
	// Number of parallel workers for markdown rendering
	MaxWorkers = 50
)

// MarkdownRenderer handles rendering JSON conversations to markdown
type MarkdownRenderer struct {
	outputPath string
}

// renderJob represents a file to be rendered
type renderJob struct {
	inputPath  string
	outputPath string
	jobType    string // "conversation" or "project"
}

// New creates a new markdown renderer instance
func New(outputPath string) *MarkdownRenderer {
	return &MarkdownRenderer{
		outputPath: outputPath,
	}
}

// RenderAll renders all conversations and projects to markdown
func (r *MarkdownRenderer) RenderAll() error {
	fmt.Println("Rendering conversations and projects to markdown...")

	// Create markdown output directories
	if err := r.createMarkdownDirectories(); err != nil {
		return fmt.Errorf("failed to create markdown directories: %w", err)
	}

	// Render Claude conversations
	if err := r.renderClaudeConversations(); err != nil {
		fmt.Printf("Warning: Claude conversation rendering failed: %v\n", err)
	}

	// Render Claude projects
	if err := r.renderClaudeProjects(); err != nil {
		fmt.Printf("Warning: Claude project rendering failed: %v\n", err)
	}

	// Render ChatGPT conversations
	if err := r.renderChatGPTConversations(); err != nil {
		fmt.Printf("Warning: ChatGPT conversation rendering failed: %v\n", err)
	}

	fmt.Println("✓ Markdown rendering completed")
	return nil
}

// createMarkdownDirectories creates the markdown output directory structure
func (r *MarkdownRenderer) createMarkdownDirectories() error {
	dirs := []string{
		"claude/chats-md",
		"claude/projects-md",
		"chatgpt/chats-md",
		"chatgpt/projects-md",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(r.outputPath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return err
		}
	}

	return nil
}

// renderClaudeConversations renders all Claude conversation JSON files to markdown using parallel processing
func (r *MarkdownRenderer) renderClaudeConversations() error {
	chatsPath := filepath.Join(r.outputPath, "claude", "chats")
	
	// Collect all conversation files
	var jobs []renderJob
	err := filepath.Walk(chatsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Generate markdown output path
		relPath, err := filepath.Rel(chatsPath, path)
		if err != nil {
			return err
		}
		
		mdPath := strings.Replace(relPath, ".json", ".md", 1)
		outputPath := filepath.Join(r.outputPath, "claude", "chats-md", mdPath)

		jobs = append(jobs, renderJob{
			inputPath:  path,
			outputPath: outputPath,
			jobType:    "conversation",
		})

		return nil
	})

	if err != nil {
		return err
	}

	// Process jobs in parallel
	return r.processJobsParallel(jobs)
}

// renderChatGPTConversations renders all ChatGPT conversation JSON files to markdown using parallel processing
func (r *MarkdownRenderer) renderChatGPTConversations() error {
	chatsPath := filepath.Join(r.outputPath, "chatgpt", "chats")
	
	// Collect all conversation files
	var jobs []renderJob
	err := filepath.Walk(chatsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Generate markdown output path
		relPath, err := filepath.Rel(chatsPath, path)
		if err != nil {
			return err
		}
		
		mdPath := strings.Replace(relPath, ".json", ".md", 1)
		outputPath := filepath.Join(r.outputPath, "chatgpt", "chats-md", mdPath)

		jobs = append(jobs, renderJob{
			inputPath:  path,
			outputPath: outputPath,
			jobType:    "conversation",
		})

		return nil
	})

	if err != nil {
		return err
	}

	// Process jobs in parallel
	return r.processJobsParallel(jobs)
}

// renderClaudeProjects renders all Claude project JSON files to markdown using parallel processing
func (r *MarkdownRenderer) renderClaudeProjects() error {
	projectsPath := filepath.Join(r.outputPath, "claude", "projects")
	
	// Collect all project files
	var jobs []renderJob
	err := filepath.Walk(projectsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, "project.json") {
			return nil
		}

		// Generate markdown output path
		projectDir := filepath.Dir(path)
		relPath, err := filepath.Rel(projectsPath, projectDir)
		if err != nil {
			return err
		}
		
		outputPath := filepath.Join(r.outputPath, "claude", "projects-md", relPath, "project.md")

		jobs = append(jobs, renderJob{
			inputPath:  path,
			outputPath: outputPath,
			jobType:    "project",
		})

		return nil
	})

	if err != nil {
		return err
	}

	// Process jobs in parallel
	return r.processJobsParallel(jobs)
}

// renderConversationToMarkdown renders a conversation to markdown format
func (r *MarkdownRenderer) renderConversationToMarkdown(conv models.Conversation, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write conversation header
	fmt.Fprintf(file, "# %s\n\n", conv.Metadata.Title)
	fmt.Fprintf(file, "**Platform:** %s  \n", conv.Metadata.Platform)
	fmt.Fprintf(file, "**Created:** %s  \n", conv.Metadata.CreatedDate.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "**Last Modified:** %s  \n", conv.Metadata.LastModified.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "**Messages:** %d  \n", conv.Metadata.MessageCount)
	if len(conv.Metadata.Participants) > 0 {
		fmt.Fprintf(file, "**Participants:** %s  \n", strings.Join(conv.Metadata.Participants, ", "))
	}
	if conv.Metadata.Project != "" {
		fmt.Fprintf(file, "**Project:** %s  \n", conv.Metadata.Project)
	}
	if len(conv.Metadata.Topics) > 0 {
		fmt.Fprintf(file, "**Topics:** %s  \n", strings.Join(conv.Metadata.Topics, ", "))
	}
	fmt.Fprintf(file, "**Has Code:** %v  \n", conv.Metadata.HasCode)
	fmt.Fprintf(file, "**Has Media:** %v  \n", conv.Metadata.HasMedia)
	fmt.Fprintf(file, "\n---\n\n")

	// Write messages
	if conv.Messages == nil || len(conv.Messages) == 0 {
		fmt.Fprintf(file, "*No messages in this conversation.*\n")
		return nil
	}

	for i, msg := range conv.Messages {
		// Determine role separator
		var roleSeparator string
		switch strings.ToLower(msg.Author) {
		case "user", "human":
			roleSeparator = ">>>user:>>>"
		case "claude", "assistant":
			roleSeparator = ">>>claude:>>>"
		case "chatgpt":
			roleSeparator = ">>>chatgpt:>>>"
		case "system":
			roleSeparator = ">>>system:>>>"
		case "tool":
			roleSeparator = ">>>tool:>>>"
		default:
			roleSeparator = fmt.Sprintf(">>>%s:>>>", strings.ToLower(msg.Author))
		}

		// Write message separator with inline timestamp
		fmt.Fprintf(file, "%s    *%s*\n\n", roleSeparator, msg.Timestamp.Format("2006-01-02 15:04:05"))

		// Write message content
		content := strings.TrimSpace(msg.Content)
		if content == "" {
			content = "*[Empty message]*"
		}

		// Format content for markdown (escape if needed, preserve code blocks)
		fmt.Fprintf(file, "%s\n", content)

		// Add spacing between messages (except for the last one)
		if i < len(conv.Messages)-1 {
			fmt.Fprintf(file, "\n")
		}
	}

	return nil
}

// renderProjectToMarkdown renders a project to markdown format
func (r *MarkdownRenderer) renderProjectToMarkdown(project models.ClaudeProject, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write project header
	fmt.Fprintf(file, "# %s\n\n", project.Name)
	fmt.Fprintf(file, "**UUID:** %s  \n", project.UUID)
	fmt.Fprintf(file, "**Created:** %s  \n", project.CreatedAt)
	fmt.Fprintf(file, "**Updated:** %s  \n", project.UpdatedAt)
	fmt.Fprintf(file, "**Documents:** %d  \n", len(project.Docs))
	fmt.Fprintf(file, "\n## Description\n\n")
	
	if project.Description != "" {
		fmt.Fprintf(file, "%s\n\n", project.Description)
	} else {
		fmt.Fprintf(file, "*No description provided.*\n\n")
	}

	// Write documents section
	if len(project.Docs) > 0 {
		fmt.Fprintf(file, "## Project Documents\n\n")
		
		for i, doc := range project.Docs {
			fmt.Fprintf(file, "### %d. %s\n\n", i+1, doc.Filename)
			if doc.CreatedAt != "" {
				fmt.Fprintf(file, "**Created:** %s  \n\n", doc.CreatedAt)
			}
			
			content := strings.TrimSpace(doc.Content)
			if content == "" {
				content = "*[Empty document]*"
			}
			
			fmt.Fprintf(file, "%s\n\n", content)
			
			if i < len(project.Docs)-1 {
				fmt.Fprintf(file, "---\n\n")
			}
		}
	}

	return nil
}

// processJobsParallel processes render jobs using a worker pool
func (r *MarkdownRenderer) processJobsParallel(jobs []renderJob) error {
	if len(jobs) == 0 {
		return nil
	}

	// Create job channel and result channel
	jobChan := make(chan renderJob, len(jobs))
	resultChan := make(chan error, len(jobs))

	// Determine number of workers (don't exceed job count)
	numWorkers := MaxWorkers
	if len(jobs) < numWorkers {
		numWorkers = len(jobs)
	}

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go r.worker(&wg, jobChan, resultChan)
	}

	// Send jobs to workers
	for _, job := range jobs {
		jobChan <- job
	}
	close(jobChan)

	// Wait for all workers to complete
	wg.Wait()
	close(resultChan)

	// Check for errors
	var errors []error
	for err := range resultChan {
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		// Log warnings but don't fail completely
		for _, err := range errors {
			fmt.Printf("Warning: markdown rendering error: %v\n", err)
		}
	}

	fmt.Printf("✓ Rendered %d files to markdown using %d workers\n", len(jobs), numWorkers)
	return nil
}

// worker processes render jobs from the job channel
func (r *MarkdownRenderer) worker(wg *sync.WaitGroup, jobChan <-chan renderJob, resultChan chan<- error) {
	defer wg.Done()

	for job := range jobChan {
		err := r.processJob(job)
		resultChan <- err
	}
}

// processJob processes a single render job
func (r *MarkdownRenderer) processJob(job renderJob) error {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(job.outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", job.outputPath, err)
	}

	switch job.jobType {
	case "conversation":
		var conv models.Conversation
		if err := r.readJSON(job.inputPath, &conv); err != nil {
			return fmt.Errorf("failed to read conversation %s: %w", job.inputPath, err)
		}
		return r.renderConversationToMarkdown(conv, job.outputPath)
	
	case "project":
		var project models.ClaudeProject
		if err := r.readJSON(job.inputPath, &project); err != nil {
			return fmt.Errorf("failed to read project %s: %w", job.inputPath, err)
		}
		return r.renderProjectToMarkdown(project, job.outputPath)
	
	default:
		return fmt.Errorf("unknown job type: %s", job.jobType)
	}
}

// readJSON reads and unmarshals a JSON file
func (r *MarkdownRenderer) readJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}