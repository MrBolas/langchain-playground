package agent

import (
	"fmt"
	"github.com/MrBolas/langchain-playground/agent/models"
	"log"
	"strings"
)

type Agent struct {
	model           string
	role            string
	roleDescription string
	maxSteps        int
	client          *models.OllamaClient
	tools           []models.Tool
}

func NewAgent(model string, role string, roleDescription string, maxSteps int, tools []models.Tool) *Agent {

	ollama := models.NewOllamaClient("http://localhost:11434/api/chat")

	return &Agent{model: model,
		maxSteps:        maxSteps,
		tools:           tools,
		role:            role,
		roleDescription: roleDescription,
		client:          ollama,
	}
}

func (a *Agent) Chat(Messages []models.Message) *models.Message {
	var haltAgent bool

	// setup request
	request := models.OllamaRequest{
		Model:    a.model,
		Messages: Messages,
		Tools:    a.tools,
		Stream:   false,
	}

	for i := 0; i < a.maxSteps; i++ {

		// call ollama
		response, err := a.client.Call(request)
		if err != nil {
			log.Printf("Error calling Ollama API: %v", err)
			return nil
		}

		if len(response.Message.ToolCall) != 0 {
			haltAgent = a.handleToolCalls(response.Message.ToolCall, &request)
		} else {
			request.Messages = append(request.Messages, models.Message{
				Role:    a.role,
				Content: response.Message.Content,
			})
		}

		if haltAgent {
			log.Printf("Agent halted")
			return &request.Messages[len(request.Messages)-2]
		}
	}

	return &request.Messages[len(request.Messages)-1]
}

func (a *Agent) Prompt(prompt string) string {
	var haltAgent bool

	// setup request
	request := models.OllamaRequest{
		Model: a.model,
		Messages: []models.Message{
			{
				Role:    a.role,
				Content: a.roleDescription},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Tools:  a.tools,
		Stream: false,
	}

	for i := 2; i < a.maxSteps; i++ {

		// call ollama
		response, err := a.client.Call(request)
		if err != nil {
			log.Printf("Error calling Ollama API: %v", err)
			return err.Error()
		}

		if len(response.Message.ToolCall) != 0 {
			haltAgent = a.handleToolCalls(response.Message.ToolCall, &request)
		} else {
			request.Messages = append(request.Messages, models.Message{
				Role:    a.role,
				Content: response.Message.Content,
			})
		}

		if haltAgent {
			break
		}
	}

	for i, message := range request.Messages {
		log.Printf("[epoch %d] %s : %s", i, message.Role, strings.ReplaceAll(message.Content, "\n", ""))
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
			log.Printf("Calling function %s with arguments %+v", tool.Function.Name, toolCall.FunctionCall.Arguments)
			result, err := safeCall(tool.Function.Call, toolCall.FunctionCall.Arguments)
			if err != nil {
				request.Messages = append(request.Messages, models.Message{
					Role:    a.role,
					Content: fmt.Sprintf("Error calling the tool %s: %v", tool.Function.Name, err),
				})
				if strings.Contains(err.Error(), "Agent halted") {
					log.Printf("Agent halted")
					return true
				}
				continue
			}
			request.Messages = append(request.Messages, models.Message{
				Role: "system",
				Content: fmt.Sprintf("function llm call %s executed with arguments %+v and result  %v. "+
					"%s please transform this result into a human message.",
					tool.Function.Name, toolCall.FunctionCall.Arguments, result, a.role),
			})
			// Call Ollama to generate a proper answer using the tool call result
			call, err := a.client.Call(*request)
			if err != nil {
				log.Printf("Error calling Ollama API: %v", err)
				return false
			}
			request.Messages = append(request.Messages, models.Message{
				Role:    a.role,
				Content: call.Message.Content,
			})
			if call.Message.Content != "" {
				return true
			}
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
