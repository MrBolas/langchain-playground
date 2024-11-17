package main

import (
	"errors"
	"github.com/MrBolas/langchain-playground/agent"
	"github.com/MrBolas/langchain-playground/agent/models"
)

func main() {

	tools := []models.Tool{
		{
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
		},
		{
			Type: "function",
			Function: models.Function{
				Name:        "stop_execution",
				Description: "Halt the agent execution, the agent purpose was fulfilled",
				Parameters:  map[string]string{},
				Call: func(args map[string]any) (any, error) {
					return nil, errors.New("Agent halted")
				},
			},
		},
	}

	description1 := "you are agent an ai agent that does one shot but writes elaborate answers. You have tools available to you: sum, stop_execution. After elaborating on the user question run function stop_execution."
	agent1 := agent.NewAgent("qwen2.5:7b", "agent", description1, 6, tools)

	//description2 := "You are agent-2 a pro communism, you want to start a conversation with agent-1. You have the tools: sum at your disposal. be very brief in your responses"
	//agent2 := agent.NewAgent("qwen2.5:7b", "agent-2", description2, 3, tools)

	agent1.Prompt("whats the sum of the first 3 fibonnaci sequence digits?")
	//agent.Prompt("call function stop_agent")

	//agent.NewConversation([]agent.Agent{*agent1, *agent2}, 10).Start()
}
