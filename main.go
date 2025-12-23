package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	openAIEndpoint = "https://api.openai.com/v1/chat/completions"
	model          = "gpt-4o-mini"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment")
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "OPENAI_API_KEY is not set")
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: chatgpt-cli <prompt>")
		os.Exit(1)
	}

	prompt := strings.Join(os.Args[1:], " ")

	reqBody := ChatCompletionRequest{
		Model: model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to encode request:", err)
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", openAIEndpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create request:", err)
		os.Exit(1)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "request failed:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "API error: %s\n%s\n", resp.Status, body)
		os.Exit(1)
	}

	var res ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		fmt.Fprintln(os.Stderr, "failed to decode response:", err)
		os.Exit(1)
	}

	if len(res.Choices) == 0 {
		fmt.Fprintln(os.Stderr, "no response from model")
		os.Exit(1)
	}

	fmt.Println(res.Choices[0].Message.Content)
}
