package ingestion

import (
	"os"
	"strings"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type Folder struct {
	Files []File
	Path  string
}

func NewFolder(path string) (Folder, error) {

	var files []File

	newFiles, err := os.ReadDir(path)
	if err != nil {
		return Folder{}, err
	}

	for _, newFile := range newFiles {
		if !newFile.IsDir() {
			files = append(files, NewFile(strings.Join([]string{path, newFile.Name()}, "/")))
		}
	}

	return Folder{
		Files: files,
		Path:  path,
	}, nil
}

func (f *Folder) Split(opts ...textsplitter.Option) []schema.Document {
	docs := []schema.Document{}
	// split files
	for _, file := range f.Files {
		docs = append(docs, file.Split()...)
	}

	return docs
}
