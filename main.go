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

	"github.com/alecthomas/kong"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type CLI struct {
	ApiKey     string   `help:"OpenAI API key"`
	Endpoint   string   `help:"OpenAI endpoint" default:"" `
	Filename   []string `arg:"" type:"existingfile" help:"PDF file to rename"`
	ImageModel string   `help:"OpenAI image model" default:"llama3.2-vision" required:""`
}

//go:embed prompt.md
var prompt string

func (c *CLI) Run() error {
	var result BottlesSchema
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		return fmt.Errorf("failed to generate schema: %w", err)
	}

	var imageMessages []openai.ChatMessagePart

	for _, imagePath := range c.Filename {
		imageFile, err := os.Open(imagePath)
		if err != nil {
			log.Fatalf("Failed to open image file: %v", err)
		}
		defer imageFile.Close()

		image, err := jpeg.Decode(imageFile)
		if err != nil {
			log.Fatalf("Failed to decode image: %v", err)
		}

		buffer := &bytes.Buffer{}
		err = jpeg.Encode(buffer, image, &jpeg.Options{Quality: 100})
		if err != nil {
			log.Fatalf("Failed to encode image: %v", err)
		}

		encodedImage := base64.StdEncoding.EncodeToString(buffer.Bytes())
		imageMessages = append(imageMessages, openai.ChatMessagePart{
			Type: "image_url",
			ImageURL: &openai.ChatMessageImageURL{
				URL:    "data:image/jpeg;base64," + encodedImage,
				Detail: openai.ImageURLDetailAuto,
			},
		})
	}

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
					Content: prompt,
				},
				{
					Role: "user",
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

	err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &result)
	if err != nil {
		log.Fatalf("Failed to unmarshal response: %v", err)
	}

	mergedBottles := mergeBottles(result.Bottles)
	mergedData, err := json.MarshalIndent(mergedBottles, "", "  ")
	if err != nil {
			return fmt.Errorf("failed to marshal merged data: %w", err)
	}

	fmt.Println(string(mergedData))

	return nil
}

func main() {
	cli := &CLI{}
	ctx := kong.Parse(cli)
	// Call the Run() method of the selected parsed command.
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
