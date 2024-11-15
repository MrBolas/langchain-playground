package main

import (
	"errors"
	"github.com/MrBolas/langchain-playground/agent"
	"github.com/MrBolas/langchain-playground/agent/models"
)

func main() {

	tools := []models.Tool{}
	tools = append(tools, models.Tool{
		Type: "function",
		Function: models.Function{
			Name:        "sum",
			Description: "Add two numbers",
			Parameters: map[string]string{
				"a": "int",
				"b": "int",
			},
			Call: func(args map[string]any) (any, error) {
				aValue, ok := args["a"]
				if !ok {
					return nil, errors.New("tool parameter 'a' does not exist")
				}
				a, ok := aValue.(float64)
				if !ok {
					return nil, errors.New("tool parameter 'a' is not a float64")
				}
				bValue, ok := args["b"]
				if !ok {
					return nil, errors.New("tool parameter 'b' does not exist")
				}
				b, ok := bValue.(float64)
				if !ok {
					return nil, errors.New("tool parameter 'b' is not a float64")
				}
				return int(a + b), nil
			},
		},
	})
	tools = append(tools, models.Tool{
		Type: "function",
		Function: models.Function{
			Name:        "stop_execution",
			Description: "Halt the agent execution, the agent purpose was fulfilled",
			Parameters:  map[string]string{},
			Call: func(args map[string]any) (any, error) {
				return nil, errors.New("Agent halted")
			},
		},
	})

	agent := agent.NewAgent("qwen2.5:7b", 10, tools)

	agent.Prompt("whats the fibonnaci sequence?")
	//agent.Prompt("call function stop_agent")
}
