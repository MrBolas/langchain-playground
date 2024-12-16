package agent

import (
	"github.com/MrBolas/langchain-playground/agent/models"
	"log"
)

type Conversation struct {
	topic        string
	participants []Agent
	rounds       int
}

func NewConversation(topic string, participants []Agent, rounds int) *Conversation {
	return &Conversation{topic: topic, participants: participants, rounds: rounds}
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
		Role:    "system",
		Content: c.topic,
	})

	log.Printf("Starting conversation with %d participants", len(c.participants))
	log.Printf("Rounds: %d", c.rounds)
	//log.Printf("Messages: %v", messages)

	for i := 0; i < c.rounds; i++ {
		for _, participant := range c.participants {
			// send messages to participant
			newMessage := participant.Chat(messages)
			messages = append(messages, *newMessage)
			log.Printf("[Round %d][Participant %s]: %s", i, participant.role, newMessage.Content)
		}
	}
}
