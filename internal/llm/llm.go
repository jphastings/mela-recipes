package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
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
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature"`
	Stream      bool      `json:"stream"`
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

type RawJSON string

func (j RawJSON) MarshalJSON() ([]byte, error) {
	return []byte(j), nil
}

type MultiAnswer struct {
	Key    string
	Answer string
	Err    error
}

var ErrWasIgnored = errors.New("the ignore prompt returned 'yes'")

func (c *Connection) MultiQuery(systemPrompt, ignorePrompt, input string, prompts map[string]string) (<-chan MultiAnswer, error) {
	ctx := context.Background()

	ans, err := c.Query(ctx, systemPrompt, input, ignorePrompt)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(strings.ToLower(ans), "yes") {
		return nil, ErrWasIgnored
	}

	answers := make(chan MultiAnswer)
	var wg sync.WaitGroup

	for key, prompt := range prompts {
		wg.Add(1)
		go func(key, prompt string) {
			defer wg.Done()

			ans, err := c.Query(ctx, systemPrompt, input, prompt)
			answers <- MultiAnswer{
				Key:    key,
				Answer: ans,
				Err:    err,
			}
		}(key, prompt)
	}

	go func() {
		wg.Wait()
		close(answers)
	}()

	return answers, nil
}

func (c *Connection) Query(ctx context.Context, systemPrompt, input, prompt string) (string, error) {
	req := request{
		Model: c.Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: input},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.7,
		Stream:      false,
	}

	reqEnc, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	res, err := httpClient.Post(c.Endpoint, "application/json", bytes.NewReader(reqEnc))
	if err != nil {
		return "", err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var answer response
	if err := json.Unmarshal(data, &answer); err != nil {
		return "", fmt.Errorf("couldn't decode JSON response: %w", err)
	}

	if len(answer.Choices) == 0 {
		return "", fmt.Errorf("the LLM returned zero response messages")
	}

	responseText := answer.Choices[0].Message.Content

	return responseText, nil
}
