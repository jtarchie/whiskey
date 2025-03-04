package main

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"log/slog"

	"github.com/philippgille/chromem-go"
	"github.com/sashabaranov/go-openai"
)

type Organize struct {
	ApiKey         string   `help:"OpenAI API key"`
	Endpoint       string   `help:"OpenAI endpoint" default:"" `
	Filename       []string `arg:"" type:"existingfile" help:"Image file to rename"`
	ImageModel     string   `help:"OpenAI image model" default:"llama3.2-vision" required:""`
	EmbeddingModel string   `help:"OpenAI embedding model" default:"llama3.2" required:""`
	EmbeddingDB    string   `help:"embedding filename" default:"embedding.db" required:"" type:"path"`
}

//go:embed prompts/organize.md
var organizePrompt string

func (o *Organize) Run() error {
	slog.Info("organize_run_start")

	db, err := chromem.NewPersistentDB(o.EmbeddingDB, false)
	if err != nil {
		return fmt.Errorf("failed to create db: %w", err)
	}
	slog.Info("created_persistent_db", "endpoint", o.Endpoint, "embedding_model", o.EmbeddingModel)

	o.Endpoint = strings.TrimPrefix(o.Endpoint, "/")

	collection, err := db.GetOrCreateCollection("filenames", nil, chromem.NewEmbeddingFuncOpenAICompat(
		o.Endpoint,
		o.ApiKey,
		o.EmbeddingModel,
		nil,
	))
	if err != nil {
		return fmt.Errorf("failed to get or create collection: %w", err)
	}
	slog.Info("got_or_created_collection", "collection", collection)

	config := openai.DefaultConfig(o.ApiKey)
	if o.Endpoint != "" {
		config.BaseURL = o.Endpoint
	}
	client := openai.NewClientWithConfig(config)

	for index, filename := range o.Filename {
		slog.Info("processing_file", "filename", filename, "index", index, "total", len(o.Filename))

		_, err = collection.GetByID(context.Background(), filename)
		if err == nil {
			slog.Info("skipping_file_already_processed", "filename", filename)
			continue
		}

		imageMessages, err := imagesAsMessages([]string{filename})
		if err != nil {
			return fmt.Errorf("failed to create image messages: %w", err)
		}
		slog.Info("created_image_messages", "count", len(imageMessages))

		response, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: o.ImageModel,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    "system",
						Content: organizePrompt,
					},
					{
						Role:         "user",
						MultiContent: imageMessages,
					},
				},
			},
		)
		if err != nil {
			return fmt.Errorf("failed to create chat completion: %w", err)
		}

		message := response.Choices[0].Message.Content
		slog.Info("received_response_from_openai", "message", message)

		err = collection.AddDocument(context.Background(), chromem.Document{
			ID:      filename,
			Content: response.Choices[0].Message.Content,
		})
		if err != nil {
			return fmt.Errorf("failed to add document: %w", err)
		}
		slog.Info("added_document_to_collection", "filename", filename)
	}

	for _, filename := range o.Filename {
		if !strings.Contains(filename, "IMG_3675") {
			continue
		}

		doc, err := collection.GetByID(context.Background(), filename)
		if err != nil {
			return fmt.Errorf("failed to get document: %w", err)
		}
		slog.Info("got_document", "filename", filename)

		results, err := collection.QueryEmbedding(context.Background(), doc.Embedding, 3, nil, nil)
		if err != nil {
			return fmt.Errorf("could not query embedding: %w", err)
		}

		slog.Info("queried_embedding", "filename", filename, "results", len(results), "contents", doc.Content)
		for _, result := range results {
			if result.ID == filename {
				continue
			}

			slog.Info("result", "filename", result.ID, "score", result.Similarity, "contents", result.Content)
		}
	}

	slog.Info("organize_run_end")
	return nil
}
