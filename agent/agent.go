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
	halt            bool
}

type IAgent interface {
	Plan(input string, request *models.OllamaRequest) (*models.OllamaRequest, error)
	Execute(request *models.OllamaRequest) (*models.OllamaResponse, error)
	Monitor(response *models.OllamaResponse, request *models.OllamaRequest) (bool, error)
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

func (a *Agent) Plan(input string, request *models.OllamaRequest) (*models.OllamaRequest, error) {
	// The planning phase: Prepare the Ollama request
	request.Messages = append(request.Messages, models.Message{
		Role:    "system",
		Content: "Generate a plan for the user request. " + input,
	})

	// make the call to the llm
	rsp, err := a.client.Call(request)
	if err != nil {
		return nil, err
	}

	// Append the response to the conversation
	request.Messages = append(request.Messages, models.Message{
		Role:    a.role,
		Content: rsp.Message.Content,
	})
	return request, nil
}

func (a *Agent) Execute(request *models.OllamaRequest) (*models.OllamaResponse, error) {
	// The execution phase: Call Ollama API
	request.Messages = append(request.Messages, models.Message{
		Role:    a.role,
		Content: "Execute the plan",
	})

	response, err := a.client.Call(request)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}
	return response, nil
}

func (a *Agent) Monitor(response *models.OllamaResponse, request *models.OllamaRequest) (bool, error) {
	if len(response.Message.ToolCall) != 0 {
		// Handle tool calls
		toolState := a.handleTools(response.Message.ToolCall, request)
		request.Messages = append(request.Messages, models.Message{
			Role:    "tools",
			Content: fmt.Sprintf("Tool state: %+v", toolState),
		})

		// Check if all required tools have been used and results processed
		if _, ok := toolState["weather_forecast"]; ok {
			a.halt = true // Halt if weather forecast is fulfilled
		}

		return !a.halt, nil
	}

	// Append response content to the conversation
	request.Messages = append(request.Messages, models.Message{
		Role:    a.role,
		Content: response.Message.Content,
	})

	// Check if the response fulfills the user's request
	if strings.Contains(response.Message.Content, "The weather in") {
		a.halt = true // Stop if a summary message is detected
	}

	return !a.halt, nil
}

func (a *Agent) Chat(Messages []models.Message) *models.Message {
	request := &models.OllamaRequest{
		Model:    a.model,
		Messages: Messages,
		Tools:    a.tools,
		Stream:   false,
	}

	for i := 0; i < a.maxSteps; i++ {
		log.Printf("Epoch %d: Starting PEM cycle", i+1)

		// Plan Phase
		var err error
		request, err = a.Plan(request.Messages[len(request.Messages)-1].Content, request)
		if err != nil {
			log.Printf("Planning failed: %v", err)
			return nil
		}

		// Execute Phase
		response, err := a.Execute(request)
		if err != nil {
			log.Printf("Execution failed: %v", err)
			return nil
		}

		// Monitor Phase
		continueExecution, err := a.Monitor(response, request)
		if err != nil {
			log.Printf("Monitoring failed: %v", err)
			return nil
		}
		if !continueExecution {
			log.Printf("Halting agent as per monitor decision")
			break
		}
	}

	return &request.Messages[len(request.Messages)-1]
}

func (a *Agent) Prompt(prompt string) string {
	// Initialize the request with the initial prompt
	request := &models.OllamaRequest{
		Model: a.model,
		Messages: []models.Message{
			{
				Role:    a.role,
				Content: a.roleDescription,
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Tools:  a.tools,
		Stream: false,
	}

	for i := 0; i < a.maxSteps; i++ {
		log.Printf("Epoch %d: Starting PEM cycle for prompt", i+1)

		// Plan Phase
		plannedRequest, err := a.Plan(request.Messages[len(request.Messages)-1].Content, request)
		if err != nil {
			log.Printf("Planning failed: %v", err)
			return fmt.Sprintf("Error: %v", err)
		}

		// Execute Phase
		response, err := a.Execute(plannedRequest)
		if err != nil {
			log.Printf("Execution failed: %v", err)
			return fmt.Sprintf("Error: %v", err)
		}

		// Monitor Phase
		continueExecution, err := a.Monitor(response, request)
		if err != nil {
			log.Printf("Monitoring failed: %v", err)
			return fmt.Sprintf("Error: %v", err)
		}
		if !continueExecution {
			log.Printf("Halting agent as per monitor decision")
			break
		}
	}

	// Log all messages for debugging
	for i, message := range request.Messages {
		log.Printf("[epoch %d] %s: %s", i, message.Role, strings.ReplaceAll(message.Content, "\n", ""))
	}

	// Return the last message from the agent
	return request.Messages[len(request.Messages)-1].Content
}

func (a *Agent) handleTools(toolCalls []models.ToolCall, request *models.OllamaRequest) map[string]any {

	toolsState := make(map[string]any)
	toolMap := make(map[string]models.Tool)
	for _, tool := range a.tools {
		toolMap[tool.Function.Name] = tool
	}

	for _, toolCall := range toolCalls {
		var tool models.Tool
		var exists bool
		if tool, exists = toolMap[toolCall.FunctionCall.Name]; !exists {
			request.Messages = append(request.Messages, models.Message{
				Role:    a.role,
				Content: fmt.Sprintf("Error: tool %s does not exist", toolCall.FunctionCall.Name),
			})
			continue
		}

		// call tool
		result, err := tool.Function.Call(toolCall.FunctionCall.Arguments)
		if err != nil {
			request.Messages = append(request.Messages, models.Message{
				Role:    a.role,
				Content: fmt.Sprintf("Error calling the tool %s: %v", tool.Function.Name, err),
			})
			continue
		}

		toolsState[tool.Function.Name] = result
		if result == "halt" {
			a.halt = true
		}
		log.Printf("%+v", toolsState)
	}
	return toolsState
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
			call, err := a.client.Call(request)
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
