package ingestion

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type URL struct {
	URL      string
	Contents []string
	Filename string
	Type     string
}

func NewURL(url string) (URL, error) {

	pathElems := strings.Split(url, "/")
	filename := pathElems[len(pathElems)-1]
	filetype := strings.Split(filename, ".")[1]

	// fetch the URL
	resp, err := http.Get(url)
	if err != nil {
		return URL{}, err
	}

	//We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return URL{}, err
	}

	return URL{
		URL:      url,
		Contents: []string{string(body)},
		Filename: filename,
		Type:     filetype,
	}, nil
}

func (u *URL) Split(opts ...textsplitter.Option) []schema.Document {

	metadata := []map[string]any{make(map[string]any)}
	metadata[0]["url"] = u.URL

	var ts textsplitter.TextSplitter

	switch u.Type {
	//case "md":
	//	ts = textsplitter.NewMarkdownTextSplitter(opts...)
	case "token":
		ts = textsplitter.NewTokenSplitter(opts...)

	default:
		ts = textsplitter.NewRecursiveCharacter(opts...)
	}

	// split chunks
	documents, err := textsplitter.CreateDocuments(ts, u.Contents, metadata)
	if err != nil {
		log.Fatalf("Failed to create documents from %s with error %e", u.URL, err)
	}

	return documents
}
