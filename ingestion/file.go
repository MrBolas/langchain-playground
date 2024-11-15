package ingestion

import (
	"bytes"
	"log"
	"os"
	"strings"

	"github.com/dslipak/pdf"
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

	pathElems := strings.Split(path, "/")
	filename := pathElems[len(pathElems)-1]
	filetype := strings.Split(filename, ".")[1]

	data, err := loadFile(path, filetype)
	if err != nil {
		log.Fatalf("Failed to load file %s with error %e", path, err)
	}

	return File{
		Contents: []string{data},
		Path:     path,
		Type:     filetype,
		Filename: filename,
	}
}

func loadFile(path string, fileType string) (string, error) {

	var data string

	switch fileType {
	case "pdf":
		r, err := pdf.Open(path)
		if err != nil {
			return "", err
		}

		var buf bytes.Buffer
		b, err := r.GetPlainText()
		if err != nil {
			return "", err
		}
		buf.ReadFrom(b)
		data = buf.String()

		log.Printf("PDF text: %s\n", data)

	default:
		dat, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		data = string(dat)
	}

	return data, nil
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
