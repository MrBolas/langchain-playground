package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/textsplitter"

	"github.com/MrBolas/langchain-playground/constants"
	"github.com/MrBolas/langchain-playground/ingestion"
	"github.com/MrBolas/langchain-playground/repository"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
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

	cs, ok := os.LookupEnv("CHUNK_SIZE")
	if !ok {
		log.Panic("CHUNK_SIZE env var not set")
	}
	chunkSize, err := strconv.Atoi(cs)
	if err != nil {
		panic(err)
	}

	co, ok := os.LookupEnv("CHUNK_OVERLAP")
	if !ok {
		log.Panic("CHUNK_OVERLAP env var not set")
	}
	chunkOverlap, err := strconv.Atoi(co)
	if err != nil {
		panic(err)
	}

	//load vector store
	store := repository.NewChromaStore(col, embbedding, chromaUrl)

	// interactiveMode := flag.Bool("i", false, "interactive mode")
	// targetId := flag.String("t", "", "target id")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("----------Ingestion Shell-----------")

	for {
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		// help
		if strings.HasPrefix(text, "help") {
			fmt.Printf(constants.HelpMessage)
			continue
		}

		// by filepath
		if strings.HasPrefix(text, "file") {
			arguments, err := SanitizeInputs(text)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			file := ingestion.NewFile(arguments[0])

			docs := file.Split(
				textsplitter.WithChunkSize(chunkSize),
				textsplitter.WithChunkOverlap(chunkOverlap))

			ids, err := store.AddDocuments(context.Background(), docs)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			fmt.Printf("Added documents with ids: %v\n", ids)
			continue
		}

		// by folder
		if strings.HasPrefix(text, "folder") {
			arguments, err := SanitizeInputs(text)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			folder, err := ingestion.NewFolder(arguments[0])
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			docs := folder.Split(
				textsplitter.WithChunkSize(chunkSize),
				textsplitter.WithChunkOverlap(chunkOverlap))

			ids, err := store.AddDocuments(context.Background(), docs)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			fmt.Printf("Added documents with ids: %v\n", ids)
			continue
		}

		// by URL
		if strings.HasPrefix(text, "url") {
			arguments, err := SanitizeInputs(text)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			url, err := ingestion.NewURL(arguments[0])
			if err != nil {
				fmt.Println(err.Error())
				continue
			}

			docs := url.Split(
				textsplitter.WithChunkSize(chunkSize),
				textsplitter.WithChunkOverlap(chunkOverlap))

			ids, err := store.AddDocuments(context.Background(), docs)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			fmt.Printf("added documents: %+v\n", docs)
			fmt.Printf("Added documents with ids: %v\n", ids)
			continue
		}

		fmt.Println("Unknown command")
	}
}

func SanitizeInputs(command string) ([]string, error) {

	arguments := strings.Split(command, " ")
	if len(arguments) < 2 {
		return []string{}, errors.New("not enough arguments")
	}

	return arguments[1:], nil
}
