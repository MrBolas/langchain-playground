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

	log.Printf("[epoch 0] %s : %s", request.Messages[0].Role, request.Messages[0].Content)
	log.Printf("[epoch 1] %s : %s", request.Messages[1].Role, request.Messages[1].Content)

	ollama := models.NewOllamaClient("http://localhost:11434/api/chat")

	for i := 2; i < a.maxSteps; i++ {

		// call ollama
		response, err := ollama.Call(request)
		if err != nil {
			log.Printf("Error calling Ollama API: %v", err)
			return err.Error()
		}

		if len(response.Message.ToolCall) != 0 {
			toolCallMap := make(map[string]models.ToolCall)
			toolCalls := response.Message.ToolCall
			for _, toolCall := range toolCalls {
				toolCallMap[toolCall.FunctionCall.Name] = toolCall
			}
			for _, tool := range a.tools {
				if toolCall, exists := toolCallMap[tool.Function.Name]; exists {
					log.Printf("Function called: %s with arguments: %v", tool.Function.Name, toolCall.FunctionCall.Arguments)
					result, err := safeCall(tool.Function.Call, toolCall.FunctionCall.Arguments)
					if err != nil {
						request.Messages = append(request.Messages, models.Message{
							Role:    "agent",
							Content: fmt.Sprintf("Error calling the tool %s: %v", tool.Function.Name, err),
						})
						if strings.Contains(err.Error(), "Agent halted") {
							return "Agent halted"
						}
						continue
					}
					request.Messages = append(request.Messages, models.Message{
						Role:    "agent",
						Content: fmt.Sprintf("function %s executed with result  %v", tool.Function.Name, result),
					})
					continue
				}
			}
		} else {
			request.Messages = append(request.Messages, models.Message{
				Role:    "agent",
				Content: response.Message.Content,
			})
		}

		log.Printf("[epoch %d] %s : %s", i, request.Messages[i].Role,
			strings.ReplaceAll(request.Messages[i].Content, "\n", ""))
	}

	return prompt
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
