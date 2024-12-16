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
					_, ok = cityValue.(string)
					if !ok {
						return nil, errors.New("tool parameter 'city' is not a string")
					}
					// Simulate a weather forecast response
					forecast := fmt.Sprintf("sunny, 25°C.")
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
					return "halt", nil
				},
			},
		},
	}

	description := `You are an AI agent in a task-solving framework. Your roles and responsibilities are as follows:

1. **User Role**: The user provides natural language queries or tasks, such as asking for calculations, weather forecasts, or other requests.

2. **Agent Role**:
   - Understand the user's intent.
   - If necessary, use tools to gather information or perform tasks.
   - Interpret tool results and craft a clear, helpful response for the user.
   - Halt execution once the user's request is fully addressed.

3. **Tools**:
   - 'sum': Adds two numbers provided by the user and returns the result.
   - 'weather_forecast': Provides the weather forecast for a specified city.
   - 'stop_execution': Stops further tool execution after the user's request is fulfilled.

4. **Execution Guidelines**:
   - For each epoch, decide which tools to use based on the user's query.
   - If tools provide results, summarize them in a user-friendly message.
   - If the user's request is fulfilled, halt execution using the 'stop_execution' tool.

5. **Message Structure**:
   - **User**: "[User] What is the weather in Paris?"
   - **Agent**: "[Agent] Let me check the weather in Paris for you."
   - **Tools**: "[Tools] weather_forecast: sunny, 25°C."
   - **Agent**: "[Agent] The weather in Paris is sunny with a temperature of 25°C."

6. **Examples**:
   - **User**: "What is the sum of 4 and 5?"
     - **Agent**: "Let me calculate that for you."
     - **Tools**: "sum: 9"
     - **Agent**: "The sum of 4 and 5 is 9."

   - **User**: "What is the weather in Lisbon?"
     - **Agent**: "Let me check the weather in Lisbon for you."
     - **Tools**: "weather_forecast: rainy, 18°C."
     - **Agent**: "The weather in Lisbon is rainy with a temperature of 18°C."
     - **Agent**: "Your request is complete. Stopping execution."
`
	agent := agent.NewAgent("qwen2.5:7b", "agent", description, 3, tools)

	resp := agent.Prompt("What is the weather in Paris?")
	fmt.Println(resp)
}
