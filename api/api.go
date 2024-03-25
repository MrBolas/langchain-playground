package api

import (
	"log"
	"net/http"

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

		ollamaLLM, err := ollama.New(ollama.WithModel("gemma:2b"))
		if err != nil {
			log.Fatal(err)
		}

		docs, errSs := store.SimilaritySearch(ctx, req.Prompt, 2)
		if errSs != nil {
			log.Fatalf("query: %v\n", errSs)
		}

		content := []llms.MessageContent{
			llms.TextParts(schema.ChatMessageTypeSystem, "You are an assistant. Answer questions."),
			llms.TextParts(schema.ChatMessageTypeSystem, docs[0].PageContent),
			llms.TextParts(schema.ChatMessageTypeSystem, docs[1].PageContent),
			llms.TextParts(schema.ChatMessageTypeHuman, req.Prompt),
		}

		completion, err := ollamaLLM.GenerateContent(ctx, content)
		if err != nil {
			log.Fatalf("GenerateContent: %v\n", err)
		}

		return c.JSON(http.StatusOK, completion)
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
