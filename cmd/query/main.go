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

	col, ok := os.LookupEnv("COLLECTION")
	if !ok {
		log.Panic("COLLECTION env var not set")
	}

	embbedding, ok := os.LookupEnv("EMBBEDDING_MODEL")
	if !ok {
		log.Panic("EMBBEDDING_MODEL env var not set")
	}

	chromaUrl, ok := os.LookupEnv("CHROMA_URL")
	if !ok {
		log.Panic("CHROMA_URL env var not set")
	}

	prompt := "who is Max?"

	ctx := context.Background()

	ollamaLLM, err := ollama.New(ollama.WithModel("gemma:2b"))
	if err != nil {
		log.Fatal(err)
	}

	ollamaEmbeddingsLLM, err := ollama.New(ollama.WithModel(embbedding))
	if err != nil {
		log.Fatal(err)
	}

	ollamaEmbeder, err := embeddings.NewEmbedder(ollamaEmbeddingsLLM)
	if err != nil {
		log.Fatal(err)
	}

	store, errNs := chroma.New(
		chroma.WithChromaURL(chromaUrl),
		chroma.WithEmbedder(ollamaEmbeder),
		chroma.WithDistanceFunction(types.COSINE),
		chroma.WithNameSpace(col),
	)
	if errNs != nil {
		log.Fatalf("new: %v\n", errNs)
	}

	docs, errSs := store.SimilaritySearch(ctx, prompt, 2)
	if errSs != nil {
		log.Fatalf("query: %v\n", errSs)
	}

	content := []llms.MessageContent{
		llms.TextParts(schema.ChatMessageTypeSystem, "You are an assistant. Given a system context answer the human query. The system context is a set of documents. The human query is a question. The answer is a completion of the human query based"),
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
