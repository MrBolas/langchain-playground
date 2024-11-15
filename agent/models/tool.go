package models

// for requests
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string                            `json:"name"`
	Description string                            `json:"description"`
	Parameters  map[string]string                 `json:"parameters"`
	Arguments   map[string]int                    `json:"arguments"`
	Call        func(map[string]any) (any, error) `json:"-"`
}

// for responses
type FunctionCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type ToolCall struct {
	Type         string       `json:"type"`
	FunctionCall FunctionCall `json:"function"`
}
