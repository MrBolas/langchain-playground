package main

import (
	"errors"
	"fmt"
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
				Name:        "weather_forecast",
				Description: "Provide a weather forecast for a given city",
				Parameters: map[string]string{
					"city": "string",
				},
				Call: func(args map[string]any) (any, error) {
					cityValue, ok := args["city"]
					if !ok {
						return nil, errors.New("tool parameter 'city' does not exist")
					}
					city, ok := cityValue.(string)
					if !ok {
						return nil, errors.New("tool parameter 'city' is not a string")
					}
					// Simulate a weather forecast response
					forecast := fmt.Sprintf("The weather in %s is sunny with a high of 25°C.", city)
					return forecast, nil
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

	var toolList []string
	for _, tool := range tools {
		toolList = append(toolList, tool.Function.Name)
	}

	/*
		description1 := "You are agent an AI agent that does one shot but writes elaborate answers. " +
			"You have tools available to you: " + strings.Join(toolList, ", ") + ". " +
			"After elaborating on a function call call stop_execution to halt the agent."
		agent1 := agent.NewAgent("qwen2.5:7b", "agent", description1, 5, tools)
	*/
	description2 :=
		`You are the Event Planner AI. Your task is to handle event logistics such as venue selection, budget allocation, and scheduling.  
    - Respond only to the Marketing Strategist AI's feedback or questions.  
	- Start the discussion with a bulletpoint list of the topics to be discussed.
	- Iterate over the list and add solutions as you go.
    - Do not suggest marketing strategies; leave this to the Marketing Strategist AI.  
    - Wait for input before making further adjustments.
    - Example: 'Based on your feedback, I suggest we allocate $1000 to venue rental and $500 for catering. What do you think?'

	Wait for the Marketing Strategist AI to respond before continuing.`
	agent2 := agent.NewAgent("qwen2.5:7b", "event_planner", description2, 1, []models.Tool{})

	description3 := `You are the Marketing Strategist AI. Your task is to handle event promotion, outreach, and RSVPs.  
    - Respond only to the Event Planner AI's proposals or questions.  
    - Do not suggest logistical plans like venue selection or budgets; leave this to the Event Planner AI.  
    - Example: 'The park venue works well for family engagement. I recommend a $300 social media ad campaign to promote it.`
	agent3 := agent.NewAgent("qwen2.5:7b", "marketing_strategist", description3, 1, []models.Tool{})

	//agent1.Prompt("whats the weather like in faro, Portugal?")
	//agent.Prompt("call function stop_agent")

	agent.NewConversation([]agent.Agent{*agent2, *agent3}, 10).Start()
}
