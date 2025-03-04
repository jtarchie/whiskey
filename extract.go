package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"log/slog"
)

type Extract struct {
	ApiKey     string   `help:"OpenAI API key"`
	Endpoint   string   `help:"OpenAI endpoint" default:"" `
	Filename   []string `arg:"" type:"existingfile" help:"Image file to rename"`
	ImageModel string   `help:"OpenAI image model" default:"llama3.2-vision" required:""`
}

//go:embed prompts/extract.md
var extractPrompt string

func (c *Extract) Run() error {
	slog.Info("extract_run_start")

	var result BottlesSchema
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		return fmt.Errorf("failed to generate schema: %w", err)
	}
	slog.Info("generated_json_schema", "schema", schema)

	imageMessages, err := imagesAsMessages(c.Filename)
	if err != nil {
		return fmt.Errorf("failed to create image messages: %w", err)
	}
	slog.Info("created_image_messages", "count", len(imageMessages))

	config := openai.DefaultConfig(c.ApiKey)
	if c.Endpoint != "" {
		config.BaseURL = c.Endpoint
	}
	client := openai.NewClientWithConfig(config)

	response, err := client.CreateChatCompletion(
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
	slog.Info("received_response_from_openai", "response", response)

	err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	slog.Info("unmarshaled_response", "result", result)

	mergedBottles := mergeBottles(result.Bottles)
	mergedData, err := json.MarshalIndent(mergedBottles, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal merged data: %w", err)
	}
	slog.Info("merged_bottles_data", "mergedData", string(mergedData))

	fmt.Println(string(mergedData))

	slog.Info("extract_run_end")
	return nil
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
