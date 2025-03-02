package main

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/philippgille/chromem-go"
	"github.com/sashabaranov/go-openai"
)

type Organize struct {
	ApiKey         string   `help:"OpenAI API key"`
	Endpoint       string   `help:"OpenAI endpoint" default:"" `
	Filename       []string `arg:"" type:"existingfile" help:"Image file to rename"`
	ImageModel     string   `help:"OpenAI image model" default:"llama3.2-vision" required:""`
	EmbeddingModel string   `help:"OpenAI embedding model" default:"nomic-embed-text" required:""`
	EmbeddingDB    string   `help:"embedding filename" default:"embedding.db" required:""`
}

//go:embed prompts/organize.md
var organizePrompt string

func (o *Organize) Run() error {
	db, err := chromem.NewPersistentDB("./db", false)
	if err != nil {
		return fmt.Errorf("failed to create db: %w", err)
	}

	normalized := true
	collection, err := db.GetOrCreateCollection("filenames", nil, chromem.NewEmbeddingFuncOpenAICompat(
		o.Endpoint,
		o.ApiKey,
		o.EmbeddingModel,
		&normalized,
	))
	if err != nil {
		return fmt.Errorf("failed to get or create collection: %w", err)
	}

	config := openai.DefaultConfig(o.ApiKey)

	if o.Endpoint != "" {
		config.BaseURL = o.Endpoint
	}
	client := openai.NewClientWithConfig(config)

	for _, filename := range o.Filename {
			
		imageMessages, err := imagesAsMessages([]string{filename})
		if err != nil {
			return fmt.Errorf("failed to create image messages: %w", err)
		}

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

		err = collection.AddDocument(context.Background(), chromem.Document{
			ID: filename,
			Content: response.Choices[0].Message.Content,
		})
		if err != nil {
			return fmt.Errorf("failed to add document: %w", err)
		}
	}

	return nil
}
