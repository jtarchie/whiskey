package main

import (
	"bytes"
	"context"
	"crypto/sha512"
	"database/sql"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"log/slog"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type Extract struct {
	ApiKey     string   `help:"OpenAI API key"`
	Endpoint   string   `help:"OpenAI endpoint" default:"" `
	Filenames  []string `arg:"" type:"existingfile" help:"Image file to rename"`
	ImageModel string   `help:"OpenAI image model" default:"llama3.2-vision" required:""`
	Database   string   `help:"SQLite database file" default:"whiskey.db"`
}

//go:embed prompts/extract.md
var extractPrompt string

func (c *Extract) Run() error {
	slog.Info("extract_run_start")

	// Initialize the database
	db, err := setupDatabase(c.Database)
	if err != nil {
		return fmt.Errorf("failed to set up database: %w", err)
	}
	defer db.Close()
	slog.Info("database_initialized", "database", c.Database)

	var result BottlesSchema
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		return fmt.Errorf("failed to generate schema: %w", err)
	}
	slog.Info("generated_json_schema", "schema", schema)

	config := openai.DefaultConfig(c.ApiKey)
	if c.Endpoint != "" {
		config.BaseURL = c.Endpoint
	}
	openAIClient := openai.NewClientWithConfig(config)

	for index, filename := range c.Filenames {
		slog.Info("processing_file", "filename", filename, "index", index, "total", len(c.Filenames))

		// Calculate SHA512 of image
		imageHash, err := calculateImageHash(filename)
		if err != nil {
			return fmt.Errorf("failed to calculate image hash: %w", err)
		}
		slog.Info("image_hash_calculated", "hash", imageHash)

		imageMessages, err := imagesAsMessages([]string{filename})
		if err != nil {
			return fmt.Errorf("failed to create image messages: %w", err)
		}
		slog.Info("created_image_messages", "count", len(imageMessages))

		response, err := openAIClient.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: c.ImageModel,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    "system",
						Content: extractPrompt,
					},
					{
						Role:         "user",
						MultiContent: imageMessages,
					},
				},
				ResponseFormat: &openai.ChatCompletionResponseFormat{
					Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
					JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
						Name:   "bottles_schema",
						Schema: schema,
						Strict: true,
					},
				},
			},
		)
		if err != nil {
			log.Fatalf("Failed to process image: %v", err)
		}
		slog.Info("received_response_from_openai")

		jsonContent := strings.ReplaceAll(response.Choices[0].Message.Content, `"null"`, `null`)
		fmt.Println(jsonContent)

		// Store result in database
		now := time.Now()
		_, err = db.Exec(
			"INSERT INTO bottles (payload, image_hash, created_at, model) VALUES (?, ?, ?, ?)",
			jsonContent, imageHash, now, c.ImageModel,
		)
		if err != nil {
			return fmt.Errorf("failed to insert bottle data: %w", err)
		}
		slog.Info("bottle_data_stored_in_database", "image", filename)
	}

	slog.Info("extract_run_end")
	return nil
}

func setupDatabase(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create bottles table if it doesn't exist
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS bottles (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            payload TEXT NOT NULL,
            image_hash TEXT NOT NULL,
            created_at TEXT NOT NULL,
						model TEXT NOT NULL
        ) STRICT;
    `)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create bottles table: %w", err)
	}

	return db, nil
}

func calculateImageHash(imagePath string) (string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image for hashing: %w", err)
	}
	defer file.Close()

	hash := sha512.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func imagesAsMessages(filenames []string) ([]openai.ChatMessagePart, error) {
	slog.Info("images_as_messages_start", "filenames", filenames)

	var imageMessages []openai.ChatMessagePart

	for _, imagePath := range filenames {
		slog.Info("processing_image", "imagePath", imagePath)

		imageFile, err := os.Open(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open image: %w", err)
		}
		defer imageFile.Close()

		image, err := jpeg.Decode(imageFile)
		if err != nil {
			return nil, fmt.Errorf("failed to decode image: %w", err)
		}

		buffer := &bytes.Buffer{}
		err = jpeg.Encode(buffer, image, &jpeg.Options{Quality: 100})
		if err != nil {
			return nil, fmt.Errorf("failed to encode image: %w", err)
		}

		encodedImage := base64.StdEncoding.EncodeToString(buffer.Bytes())
		imageMessages = append(imageMessages, openai.ChatMessagePart{
			Type: "image_url",
			ImageURL: &openai.ChatMessageImageURL{
				URL:    "data:image/jpeg;base64," + encodedImage,
				Detail: openai.ImageURLDetailAuto,
			},
		})
		slog.Info("encoded_image", "encodedImageLength", len(encodedImage))
	}

	slog.Info("images_as_messages_end", "imageMessagesCount", len(imageMessages))
	return imageMessages, nil
}
