package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Connection struct {
	Endpoint string
	Model    string
	APIKey   string
}

func NewLMStudioConnection(endpoint, model string) (*Connection, error) {
	conn := &Connection{
		Endpoint: endpoint,
		Model:    model,
	}

	return conn, nil
}

type request struct {
	Model          string            `json:"model"`
	Messages       []Message         `json:"messages"`
	ResponseFormat ReqResponseFormat `json:"response_format"`
	Temperature    float32           `json:"temperature"`
	Stream         bool              `json:"stream"`
}

type response struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ReqResponseFormat struct {
	Type       string     `json:"type"`
	JSONSchema JSONSchema `json:"json_schema"`
}

type JSONSchema struct {
	Name   string  `json:"name"`
	Strict bool    `json:"strict"`
	Schema RawJSON `json:"schema"`
}

type RawJSON string

func (j RawJSON) MarshalJSON() ([]byte, error) {
	return []byte(j), nil
}

func (c *Connection) StructuredQuery(systemPrompt, userPrompt string, schema RawJSON, target any) error {
	req := request{
		Model: c.Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: ReqResponseFormat{
			Type: "json_schema",
			JSONSchema: JSONSchema{
				// Name:   "mela_recipe",
				Strict: true,
				Schema: RawJSON(schema),
			},
		},
		Temperature: 0.7,
		Stream:      false,
	}

	reqEnc, err := json.Marshal(req)
	if err != nil {
		return err
	}

	res, err := http.Post(c.Endpoint, "application/json", bytes.NewReader(reqEnc))
	if err != nil {
		return err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var answer response
	if err := json.Unmarshal(data, &answer); err != nil {
		return fmt.Errorf("couldn't decode JSON response: %w", err)
	}

	if len(answer.Choices) == 0 {
		return fmt.Errorf("the LLM returned zero response messages")
	}

	jsonText := answer.Choices[0].Message.Content

	if err := json.Unmarshal([]byte(jsonText), target); err != nil {
		return fmt.Errorf("couldn't decode JSON into provided target: %w", err)
	}

	return nil
}
