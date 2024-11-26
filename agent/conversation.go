package agent

import (
	"github.com/MrBolas/langchain-playground/agent/models"
	"log"
)

type Conversation struct {
	participants []Agent
	rounds       int
}

func NewConversation(participants []Agent, rounds int) *Conversation {
	return &Conversation{participants: participants, rounds: rounds}
}

func (c *Conversation) Start() {
	var messages []models.Message

	for _, participant := range c.participants {
		messages = append(messages, models.Message{
			Role:    participant.role,
			Content: participant.roleDescription,
		})
	}

	messages = append(messages, models.Message{
		Role: "system",
		Content: `This is a structured collaboration between two AI agents with distinct roles:  
    1. Event Planner AI: Handles logistics (venues, schedules, budgets).  
    2. Marketing Strategist AI: Handles promotions, RSVPs, and outreach strategies.  

    **Rules**:  
    - Do not repeat greetings or introductions after the first round.
    - Respond directly to the other agent’s last input.
    - Provide actionable suggestions or feedback.
	- Only respond to the other agent’s input.  
    - Do not repeat information unnecessarily.  
    - Focus on the specific question or suggestion provided by the other agent.  
    - Wait for the other agent’s input before proceeding. 

    **Goal**: 
	- Plan a community event by finalizing the venue, budget allocation, and marketing strategy in a collaborative manner.
	- The event is going to be in Lisbon, Portugal.
	- The budget is 10,000 euros.
	- The target audience is young professionals interested in technology and innovation.
	- The event will take place in 3 months.
	- The event will last for 6 hours.
	- The event will have 9 speakers.
	- The event is expected to have 300 people.

	** Conclusion**:
	- On the 10 message the conversation will end.
	- On the 10th message make a conclusion of the event planning with a list of decisions.`,
	})

	log.Printf("Starting conversation with %d participants", len(c.participants))
	log.Printf("Rounds: %d", c.rounds)
	//log.Printf("Messages: %v", messages)

	for i := 0; i < c.rounds; i++ {
		for _, participant := range c.participants {
			// send messages to participant
			newMessage := participant.Chat(messages)
			messages = append(messages, *newMessage)
			log.Printf("Participant %s: %s", participant.role, newMessage.Content)
		}
	}
}
