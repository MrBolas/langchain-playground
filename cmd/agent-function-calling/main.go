package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/duckduckgo"
	"github.com/tmc/langchaingo/tools/scraper"
	"github.com/tmc/langchaingo/tools/wikipedia"
)

func main() {

	llm, err := ollama.New(ollama.WithModel("llama3.2:3b"))
	if err != nil {
		log.Fatal(err)
	}

	wiki := wikipedia.New("Mozilla/5.0")
	if err != nil {
		log.Fatal(err)
	}

	scraper, err := scraper.New()
	if err != nil {
		log.Fatal(err)
	}

	duckduckgo, err := duckduckgo.New(3, "Mozilla/5.0")
	if err != nil {
		log.Fatal(err)
	}

	/*
		result, err := wiki.Call(context.Background(), "Olivia Wilde boyfriend")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Wikipedia Result: ", result)
	*/
	agentTools := []tools.Tool{
		duckduckgo,
		tools.Calculator{},
		wiki,
		scraper,
	}

	prompt := "Search the web for relevant stocks that prove to be a good investment for the next 5 years and give me a detailed analysis of the same. If you search the web give the urls "

	agent := agents.NewConversationalAgent(llm,
		agentTools,
		agents.WithMaxIterations(1000),
		agents.WithReturnIntermediateSteps(),
	)

	executor := agents.NewExecutor(agent,
		agents.WithMaxIterations(1000),
		agents.WithReturnIntermediateSteps())

	fmt.Printf("Agent Action: %+v\n", executor)

	answer, err := chains.Run(context.Background(), executor, prompt)
	fmt.Printf("Answer: %s\n", answer)
	log.Fatal(err)

	baseline, err := llm.Call(context.Background(), prompt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(baseline)
}
