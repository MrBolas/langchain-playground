package main

import "C"
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
					return nil, errors.New("Agent halted")
				},
			},
		},
	}

	resourceAllocator := models.Tool{
		Type: "function",
		Function: models.Function{
			Name:        "resource_allocator",
			Description: "Allocates budget to different mission phases and outputs a summary",
			Parameters: map[string]string{
				"budget": "integer",
			},
			Call: func(args map[string]any) (any, error) {
				inputValue, ok := args["budget"]
				if !ok {
					return nil, errors.New("tool parameter 'budget' does not exist")
				}
				budget, ok := inputValue.(float64)
				if !ok {
					return nil, errors.New("tool parameter 'budget' is not a string")
				}
				// Simulate a resource allocation response
				phases := []string{"Development", "Testing", "Launch"}
				allocation := map[string]float64{
					"Development": 0.5 * budget,
					"Testing":     0.3 * budget,
					"Launch":      0.2 * budget,
				}

				result := "Resource Allocation:\n"
				for _, phase := range phases {
					result += fmt.Sprintf("- %s: $%.2f\n", phase, allocation[phase])
				}
				return result, nil
			},
		},
	}
	massCalculator := models.Tool{
		Type: "function",
		Function: models.Function{
			Name:        "mass_calculator",
			Description: "Checks if the payload fits within rocket mass limits",
			Parameters: map[string]string{
				"payload_mass": "float",
			},
			Call: func(args map[string]any) (any, error) {
				payloadMassValue, ok := args["payload_mass"]
				if !ok {
					return nil, errors.New("tool parameter 'payload_mass' does not exist")
				}
				payloadMass, ok := payloadMassValue.(float64)
				if !ok {
					return nil, errors.New("tool parameter 'payload_mass' is not a float64")
				}

				// Simulate a mass calculation
				if payloadMass < 4000 {
					return fmt.Sprintf("Payload mass exceeds limit! Current: %.2f kg", payloadMass), errors.New("payload mass exceeds limit")
				}
				return fmt.Sprintf("Payload mass is within limits. Current: %.2f kg", payloadMass), nil
			},
		},
	}
	environmentSimulator := models.Tool{
		Type: "function",
		Function: models.Function{
			Name:        "environment_simulator",
			Description: "Simulates payload survivability under extreme conditions.",
			Parameters: map[string]string{
				"temperature": "float",
			},
			Call: func(args map[string]any) (any, error) {
				temperatureValue, ok := args["temperature"]
				if !ok {
					return nil, errors.New("tool parameter 'temperature' does not exist")
				}
				temperature, ok := temperatureValue.(float64)
				if !ok {
					return nil, errors.New("tool parameter 'temperature' is not a float64")
				}

				// Simulate the environment
				if temperature > -160 {
					return "It's too hot for the payload survivability.", nil
				}
				if temperature < -220 {
					return "It's too cold for the payload survivability.", nil
				}
				return "The environment on Europa is stable.", nil
			},
		},
	}

	missionPlannerTools := []models.Tool{
		resourceAllocator,
	}

	payloadSpecialistTools := []models.Tool{
		massCalculator,
		environmentSimulator,
	}

	var toolList []string
	for _, tool := range tools {
		toolList = append(toolList, tool.Function.Name)
	}

	conversationTopic := `
This is a structured collaboration between two AI agents with distinct roles:  
1. **Mission Planner AI**: Responsible for defining the mission timeline, budget allocation, and overall resource planning.  
2. **Payload Specialist AI**: Responsible for designing and ensuring the feasibility of the payload and its components.  

**Mission Context**:  
- Objective: Plan a mission to Europa (Jupiter's moon) to search for life beneath its icy surface.  
- Constraints:
  - The mission budget is $2 billion.
  - The mission timeline is 7 years from start to launch.
  - The payload must fit within the Falcon Heavy rocket's mass limit of 4,000 kg.
  - Instruments must withstand Europa's extreme environment (temperature: -160°C to -220°C, radiation: 5 Sv/hour).  

**Rules**:  
- **Mission Planner AI**:  
  - Start by defining the mission phases and using the **Resource Allocator** tool to distribute the budget.  
  - Respond only to the Payload Specialist AI's feedback or suggestions.  
  - Do not propose payload designs. Leave this to the Payload Specialist AI.  

- **Payload Specialist AI**:  
  - Propose payload components and check feasibility using tools like the **Mass Calculator** and **Environmental Simulator**.  
  - Respond only to the Mission Planner AI's timeline or resource constraints.  
  - Do not propose mission phases or budget allocations. Leave this to the Mission Planner AI.  

**Interaction Guidelines**:  
1. Respond directly to the other agent's most recent input.  
2. Use tools only when relevant and provide a summary of the tool's output.  
3. Avoid repeating information unnecessarily.  
4. Collaborate to align mission goals, ensuring technical and resource feasibility.

**Goal**:  
- Develop a detailed roadmap for the Europa mission, including:
  - Phases of development, testing, and launch.
  - Finalized payload design and mass.
  - Budget distribution.
  - Feasibility of payload under Europa's environmental conditions.

**Conclusion**:  
- On the 10th message, summarize the roadmap:
  - Mission phases and timeline.
  - Payload design specifications and feasibility.
  - Budget allocation.
  - Any key challenges and their solutions.
`

	description2 := `You are the Mission Planner AI. Your goal is to define the timeline and allocate resources for the Europa mission.`
	agent2 := agent.NewAgent("qwen2.5:7b", "mission_planner", description2, 3, missionPlannerTools)

	description3 := `You are the Payload Specialist AI. Your goal is to propose payload designs and ensure feasibility.`
	agent3 := agent.NewAgent("qwen2.5:7b", "payload_specialist", description3, 3, payloadSpecialistTools)

	agent.NewConversation(conversationTopic, []agent.Agent{*agent2, *agent3}, 10).Start()
}
