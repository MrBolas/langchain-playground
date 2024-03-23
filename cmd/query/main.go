package main

import (
	"context"
	"log"
	"os"

	"github.com/amikos-tech/chroma-go/types"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/chroma"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	prompt := "What is the process to make a spaghetti carbonara?"

	ctx := context.Background()

	ollamaLLM, err := ollama.New(ollama.WithModel("mistral"))
	if err != nil {
		log.Fatal(err)
	}

	ollamaEmbeddingsLLM, err := ollama.New(ollama.WithModel("nomic-embed-text"))
	if err != nil {
		log.Fatal(err)
	}

	ollamaEmbeder, err := embeddings.NewEmbedder(ollamaEmbeddingsLLM)
	if err != nil {
		log.Fatal(err)
	}

	store, errNs := chroma.New(
		chroma.WithChromaURL(os.Getenv("CHROMA_URL")),
		chroma.WithEmbedder(ollamaEmbeder),
		chroma.WithDistanceFunction(types.COSINE),
		chroma.WithNameSpace("80fd48be-ec21-427e-b255-19f25f35c869"),
	)
	if errNs != nil {
		log.Fatalf("new: %v\n", errNs)
	}

	docs, errSs := store.SimilaritySearch(ctx, prompt, 2)
	if errSs != nil {
		log.Fatalf("query: %v\n", errSs)
	}

	content := []llms.MessageContent{
		llms.TextParts(schema.ChatMessageTypeSystem, "You are an assistant expert."),
		llms.TextParts(schema.ChatMessageTypeSystem, docs[0].PageContent),
		llms.TextParts(schema.ChatMessageTypeSystem, docs[1].PageContent),
		llms.TextParts(schema.ChatMessageTypeHuman, prompt),
	}

	completion, err := ollamaLLM.GenerateContent(ctx, content)
	if err != nil {
		log.Fatalf("GenerateContent: %v\n", err)
	}

	log.Printf("Completion: %+v\n", completion.Choices[0].Content)

	log.Printf("SimilaritySearch: %+v\n", docs)
}
