package main

import (
	"log"
	"os"

	"github.com/amikos-tech/chroma-go/types"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/vectorstores/chroma"

	"github.com/MrBolas/langchain-playground/api"
)

func loadEnvVariables() map[string]string {

	envVariables := make(map[string]string)

	// load en variables from ".env" file
	godotenv.Load()

	chromaUrl, exists := os.LookupEnv("CHROMA_URL")
	if !exists {
		panic("CHROMA_URL undefined")
	}

	collection, exists := os.LookupEnv("COLLECTION")
	if !exists {
		panic("COLLECTION undefined")
	}

	embbeding_model, exists := os.LookupEnv("EMBBEDDING_MODEL")
	if !exists {
		panic("EMBBEDDING_MODEL undefined")
	}

	port, exists := os.LookupEnv("PORT")
	if !exists {
		panic("PORT undefined")
	}

	envVariables["CHROMA_URL"] = chromaUrl
	envVariables["COLLECTION"] = collection
	envVariables["EMBBEDDING_MODEL"] = embbeding_model
	envVariables["PORT"] = port

	return envVariables
}

func main() {
	envVar := loadEnvVariables()

	ollamaEmbeddingsLLM, err := ollama.New(ollama.WithModel(envVar["EMBBEDDING_MODEL"]))
	if err != nil {
		log.Fatal(err)
	}

	ollamaEmbeder, err := embeddings.NewEmbedder(ollamaEmbeddingsLLM)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new Chroma vector store.
	store, errNs := chroma.New(
		chroma.WithChromaURL(envVar["CHROMA_URL"]),
		chroma.WithEmbedder(ollamaEmbeder),
		chroma.WithDistanceFunction(types.COSINE),
		chroma.WithNameSpace(envVar["COLLECTION"]),
	)
	if errNs != nil {
		log.Fatalf("new: %v\n", errNs)
	}

	// Start API
	a := api.NewApi(store)
	err = a.Start(envVar["PORT"])
	if err != nil {
		log.Fatalf("unable to start echo: %v", err)
	}
}
