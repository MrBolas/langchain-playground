package agent

import (
	"fmt"
	"github.com/MrBolas/langchain-playground/agent/models"
	"log"
	"strings"
)

type Agent struct {
	model    string
	maxSteps int
	tools    []models.Tool
}

func NewAgent(model string, maxSteps int, tools []models.Tool) *Agent {
	return &Agent{model: model, maxSteps: maxSteps, tools: tools}
}

func (a *Agent) Prompt(prompt string) string {
	var haltAgent bool

	// setup request
	request := models.OllamaRequest{
		Model: a.model,
		Messages: []models.Message{
			{
				Role: "system",
				Content: "Hello, I am Agent, a concise language model agent. " +
					"I will answer your question briefly. I will answer the final answer starting by `Final Answer`."},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Tools:  a.tools,
		Stream: false,
	}

	ollama := models.NewOllamaClient("http://localhost:11434/api/chat")

	for i := 2; i < a.maxSteps; i++ {

		// call ollama
		response, err := ollama.Call(request)
		if err != nil {
			log.Printf("Error calling Ollama API: %v", err)
			return err.Error()
		}

		if len(response.Message.ToolCall) != 0 {
			haltAgent = a.handleToolCalls(response.Message.ToolCall, &request)
		} else {
			request.Messages = append(request.Messages, models.Message{
				Role:    "agent",
				Content: response.Message.Content,
			})
		}

		if haltAgent {
			break
		}
	}

	for i, message := range request.Messages {
		log.Printf("[epoch %d] %s : %s", i, message.Role, message.Content)
	}

	return prompt
}

func (a *Agent) handleToolCalls(toolCalls []models.ToolCall, request *models.OllamaRequest) bool {
	toolMap := make(map[string]models.Tool)
	for _, tool := range a.tools {
		toolMap[tool.Function.Name] = tool
	}

	for _, toolCall := range toolCalls {
		if tool, exists := toolMap[toolCall.FunctionCall.Name]; exists {
			result, err := safeCall(tool.Function.Call, toolCall.FunctionCall.Arguments)
			if err != nil {
				request.Messages = append(request.Messages, models.Message{
					Role:    "agent",
					Content: fmt.Sprintf("Error calling the tool %s: %v", tool.Function.Name, err),
				})
				if strings.Contains(err.Error(), "Agent halted") {
					return true
				}
				continue
			}
			request.Messages = append(request.Messages, models.Message{
				Role:    "agent",
				Content: fmt.Sprintf("function %s executed with result  %v", tool.Function.Name, result),
			})
		}
	}
	return false
}

func safeCall(fn func(map[string]any) (any, error), args map[string]any) (any, error) {
	var result any
	var err error
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
			result = nil
		}
	}()
	result, err = fn(args)
	return result, err
}
