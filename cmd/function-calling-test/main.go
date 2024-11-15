package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

type Function struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  map[string]string `json:"parameters"`
	Arguments   map[string]int    `json:"arguments"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type FunctionCall struct {
	Name      string         `json:"name"`
	Arguments map[string]int `json:"arguments"`
}

type ToolCall struct {
	Type         string       `json:"type"`
	FunctionCall FunctionCall `json:"function"`
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

func main() {

	// The prompt that will instruct the model to call a function
	prompt := `You are a helpful assistant that uses functions. Please use the function "sum" to add two numbers: a = 15, b = 37.`

	// Create the request payload
	requestBody := OllamaRequest{
		Model: "llama3.2:3b", // Adjust to your model
		Messages: []Message{
			{
				Role:    "assistant",
				Content: "Hello! I am a helpful assistant that uses functions.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Tools: []Tool{
			{
				Type: "function",
				Function: Function{
					Name:        "sum",
					Description: "Add two numbers",
					Parameters: map[string]string{
						"a": "int",
						"b": "int",
					},
				},
			},
		},
		Stream: false,
	}

	// Make the request to Ollama API
	responseBody, err := callOllamaAPI(requestBody)
	if err != nil {
		log.Fatalf("Error calling Ollama API: %v", err)
	}

	// Check if a function was called
	if len(responseBody.Message.ToolCall) != 0 {
		fmt.Printf("Function called: %s with arguments: %v\n", responseBody.Message.ToolCall[0].FunctionCall.Name, responseBody.Message.ToolCall[0].FunctionCall.Arguments)
		// Call the sum function with the provided arguments
		result := callSumFunction(responseBody.Message.ToolCall[0].FunctionCall.Arguments["a"], responseBody.Message.ToolCall[0].FunctionCall.Arguments["b"])
		fmt.Printf("Sum result: %d\n", result)
	} else {
		// Print the response text
		fmt.Println("Model response:", responseBody.Message.Content)
	}
}

// Function to simulate the sum function
func callSumFunction(a, b int) int {
	return a + b
}

// Function to call the Ollama API
func callOllamaAPI(requestBody OllamaRequest) (*OllamaResponse, error) {
	// Convert requestBody to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Send POST request to Ollama API
	url := "http://localhost:11434/api/chat" // Adjust the URL based on your setup
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
