package repository

import (
	"log"
	"os"

	"github.com/amikos-tech/chroma-go/types"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/vectorstores/chroma"
)

func NewChromaStore(collection string, embbederModel string) chroma.Store {

	ollamaEmbeddingsLLM, err := ollama.New(ollama.WithModel(embbederModel))
	if err != nil {
		log.Fatal(err)
	}

	ollamaEmbeder, err := embeddings.NewEmbedder(ollamaEmbeddingsLLM)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new Chroma vector store.
	store, errNs := chroma.New(
		chroma.WithChromaURL(os.Getenv("CHROMA_URL")),
		chroma.WithEmbedder(ollamaEmbeder),
		chroma.WithDistanceFunction(types.COSINE),
		chroma.WithNameSpace(collection),
	)
	if errNs != nil {
		log.Fatalf("new: %v\n", errNs)
	}

	return store
}
