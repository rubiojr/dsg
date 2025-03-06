package main

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

func sendOpenAIRequest(client *openai.Client, model, prompt string) (string, error) {
	ctx := context.Background()

	// Create chat completion request
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.2, // Lower temperature for more deterministic output
			MaxTokens:   8192,
		},
	)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices from OpenAI")
	}

	// Extract the response content
	return resp.Choices[0].Message.Content, nil
}
