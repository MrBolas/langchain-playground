package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Message struct {
	Role     string `json:"role"`
	Content  string `json:"content"`
	ToolCall []Tool `json:"tool_calls"`
}

type MessageResponse struct {
	Role     string     `json:"role"`
	Content  string     `json:"content"`
	ToolCall []ToolCall `json:"tool_calls"`
}

type OllamaRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools"`
	Stream   bool      `json:"stream"`
}

type OllamaResponse struct {
	Model     string          `json:"model"`
	CreatedAt string          `json:"created_at"`
	Message   MessageResponse `json:"message"`
}

type OllamaClient struct {
	Host string
}

type Options interface {
}
type OllamaOptions struct {
	Host string
}

func NewOllamaClient(host string) *OllamaClient {
	return &OllamaClient{Host: host}
}

func (c *OllamaClient) Call(request OllamaRequest) (*OllamaResponse, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Send POST request to Ollama API
	url := c.Host //"http://localhost:11434/api/chat" // Adjust the URL based on your setup
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response OllamaResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &response, nil
}
