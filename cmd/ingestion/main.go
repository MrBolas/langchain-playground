package main

import (
	"context"
	"log"
	"os"

	"github.com/amikos-tech/chroma-go/types"
	"github.com/gofrs/uuid"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores/chroma"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	ollamaEmbeddingsLLM, err := ollama.New(ollama.WithModel("nomic-embed-text"))
	if err != nil {
		log.Fatal(err)
	}

	ollamaEmbeder, err := embeddings.NewEmbedder(ollamaEmbeddingsLLM)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new Chroma vector store.
	collection := uuid.Must(uuid.NewV4()).String()
	store, errNs := chroma.New(
		chroma.WithChromaURL(os.Getenv("CHROMA_URL")),
		chroma.WithEmbedder(ollamaEmbeder),
		chroma.WithDistanceFunction(types.COSINE),
		chroma.WithNameSpace(collection),
	)
	if errNs != nil {
		log.Fatalf("new: %v\n", errNs)
	}

	// load file
	ts := textsplitter.NewMarkdownTextSplitter(
		textsplitter.WithChunkSize(500),
		textsplitter.WithChunkOverlap(100),
	)

	dat, err := os.ReadFile("cmd/ingestion/sample.md")
	if err != nil {
		log.Fatalf("ReadFile: %v\n", err)
	}

	metadata := []map[string]any{make(map[string]any)}
	metadata[0]["filename"] = "sample.md"
	metadata[0]["type"] = "recipe"

	// split chunks
	documents, err := textsplitter.CreateDocuments(ts, []string{string(dat)}, metadata)
	if err != nil {
		log.Fatalf("CreateDocuments: %v\n", err)
	}

	//log.Printf("document %+v\n", documents)

	// generate and store embeddings
	addedIds, errAd := store.AddDocuments(context.Background(), documents)
	if errAd != nil {
		log.Fatalf("AddDocument: %v\n", errAd)
	}

	log.Printf("Added documents with ids: %v\n to collection %s", addedIds, collection)
}
