package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/Br0ce/articleDB/pkg/encoding"
)

var (
	ErrInvalidResponse = errors.New("invalid response")
	ErrInvalidResult   = errors.New("invalid result")
	ErrBadGateway      = errors.New("bad gateway")
)

type completionDTO struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Temperature float32 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	TopP        float32 `json:"top_p"`
	FrequencyP  float32 `json:"frequency_penalty"`
	PresenceP   float32 `json:"presence_penalty"`
}

type responseDTO struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int          `json:"created"`
	Model   string       `json:"model"`
	Choices []choicesDTO `json:"choices"`
	Usage   usageDTO     `json:"usage"`
}

type choicesDTO struct {
	Text         string `json:"text"`
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason"`
}

type usageDTO struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Client struct {
	apiKey         string
	completionAddr string
	log            *zap.SugaredLogger
}

func NewClient(apiKey string, log *zap.SugaredLogger) *Client {
	return &Client{
		apiKey:         apiKey,
		completionAddr: "https://api.openai.com/v1/completions",
		log:            log,
	}
}

const gpt3TextModel = "text-davinci-003"

// Summarize uses the openAI api to perform a summarization of the given text.
func (c *Client) Summarize(ctx context.Context, text string) (string, error) {
	c.log.Infow("summarize text with openAI",
		"method", "Summarize",
		"lenText", len(text))

	if text == "" {
		return "", errors.New("could not summarize, text is empty")
	}

	dto := completionDTO{
		Model:       gpt3TextModel,
		Prompt:      fmt.Sprintf("%s\n\nTl;dr", text),
		Temperature: 1,
		MaxTokens:   220,
		TopP:        1.0,
		FrequencyP:  0.0,
		PresenceP:   1,
	}

	payload, err := encoding.EncodeToReader(dto)
	if err != nil {
		return "", err
	}

	response, err := c.execRequest(ctx, payload)
	if err != nil {
		return "", err
	}
	c.log.Debugw("response dto", "method", "responseToText", "response", response)

	return c.getResultString(response)
}

// execRequest performs the acutal post request to the openAI api and returns a responseDTO.
// To timeout the request, use an appropriate context.
// For now, there is no retrying or throttling performed.
func (c *Client) execRequest(ctx context.Context, payload io.Reader) (responseDTO, error) {
	c.log.Debugw("perform request to openAI", "method", "execRequest")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.completionAddr, payload)
	if err != nil {
		return responseDTO{}, fmt.Errorf("%s, %w", err.Error(), ErrBadGateway)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	cl := http.Client{}
	resp, err := cl.Do(req)
	if err != nil {
		return responseDTO{}, err
	}

	c.log.Debugw("response info",
		"method", "Summarize",
		"status", resp.Status,
		"headers", resp.Header)

	if resp.StatusCode >= 300 {
		return responseDTO{}, ErrBadGateway
	}

	var dto responseDTO
	err = encoding.DecodeJSON(resp.Body, &dto)
	if err != nil {
		return responseDTO{}, err
	}
	defer resp.Body.Close()

	return dto, nil
}

// getResultString extracts the result from the response and returns
// it as a string. In case multipe results are present it picks the first.
// If no result, or an empty result is found, an ErrInvalidResponse is returned.
func (c *Client) getResultString(response responseDTO) (string, error) {
	c.log.Debugw("get result from response", "method", "getResultString")

	choices := response.Choices
	if len(choices) == 0 {
		return "", ErrInvalidResponse
	}

	// In case of multiple results, we simply return the first.
	// If in the future multiple results should be used, getContentString
	// should return a slice of results.
	if len(choices) > 1 {
		c.log.Infow("openAI returned multiple results, use first result",
			"method", "getContentString",
			"resultLen", len(choices))
	}

	result := response.Choices[0].Text
	if result == "" {
		return "", ErrInvalidResult
	}

	return result, nil
}
