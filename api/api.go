package api

import (
	"log"
	"net/http"

	"github.com/MrBolas/langchain-playground/constants"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/chroma"
)

type Api struct {
	echo   *echo.Echo
	vector *chroma.Store
}

type PromptRequest struct {
	Prompt string `json:"prompt"`
}

type PromptResponse struct {
	Response string            `json:"response"`
	Context  []schema.Document `json:"context"`
}

func NewApi(store chroma.Store) *Api {

	e := echo.New()

	//middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Routes
	e.GET("/prompt", func(c echo.Context) error {

		ctx := c.Request().Context()

		req := PromptRequest{}
		c.Bind(&req)

		ollamaLLM, err := ollama.New(ollama.WithModel("gemma"),
			ollama.WithSystemPrompt(constants.SystemMessage))
		if err != nil {
			log.Fatal(err)
		}

		docs, errSs := store.SimilaritySearch(ctx, req.Prompt, 10)
		if errSs != nil {
			log.Fatalf("query: %v\n", errSs)
		}

		content := []llms.MessageContent{}
		//content = append(content, llms.TextParts(schema.ChatMessageTypeSystem, constants.SystemMessage))
		for _, doc := range docs {
			content = append(content, llms.TextParts(schema.ChatMessageTypeAI, doc.PageContent))
		}
		content = append(content, llms.TextParts(schema.ChatMessageTypeHuman, req.Prompt))

		completion, err := ollamaLLM.GenerateContent(ctx, content, llms.WithTemperature(0.2))
		if err != nil {
			log.Fatalf("GenerateContent: %v\n", err)
		}

		response := PromptResponse{
			Response: completion.Choices[0].Content,
			Context:  docs,
		}

		return c.JSON(http.StatusOK, response)
	})

	return &Api{
		echo:   e,
		vector: &store,
	}
}

func (api *Api) Start(port string) error {
	// Start server
	return api.echo.Start(":" + port)
}
