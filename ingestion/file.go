package ingestion

import (
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type IngestionSource interface {
	Split(opts ...textsplitter.Option) []schema.Document
}

type File struct {
	Contents []string
	Path     string
	Filename string
	Type     string
}

func NewFile(path string) File {

	dat, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("ReadFile: %v\n", err)
	}

	pathElems := strings.Split(path, "/")
	filename := pathElems[len(pathElems)-1]
	filetype := strings.Split(filename, ".")[1]

	return File{
		Contents: []string{string(dat)},
		Path:     path,
		Type:     filetype,
		Filename: filename,
	}
}

func (f *File) Split(opts ...textsplitter.Option) []schema.Document {

	metadata := []map[string]any{make(map[string]any)}
	metadata[0]["filename"] = f.Filename
	metadata[0]["type"] = f.Type

	var ts textsplitter.TextSplitter

	switch f.Type {
	case "md":
		ts = textsplitter.NewMarkdownTextSplitter(opts...)
	case "token":
		ts = textsplitter.NewTokenSplitter(opts...)
	default:
		ts = textsplitter.NewRecursiveCharacter(opts...)
	}

	// split chunks
	documents, err := textsplitter.CreateDocuments(ts, f.Contents, metadata)
	if err != nil {
		log.Fatalf("Failed to create documents from %s with error %e", f.Path, err)
	}

	return documents
}
