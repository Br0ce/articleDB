package openai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Br0ce/articleDB/pkg/article"
	"github.com/Br0ce/articleDB/pkg/encoding"
	"github.com/Br0ce/articleDB/pkg/request"
)

var (
	ErrInvalidResponse = errors.New("invalid response")
	ErrInvalidResult   = errors.New("invalid result")
	ErrBadGateway      = errors.New("bad gateway")
)

const (
	gpt3TextModel = "text-davinci-003"
	nerPrompt     = "List named entities with entity type person, type location and type organisation in the text. Return a json"
	sumPrompt     = "Tl;dr"
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

type nerDTO struct {
	Person       []string `json:"Person"`
	Location     []string `json:"Location"`
	Organisation []string `json:"Organization"`
}

type Client struct {
	apiKey         string
	completionAddr string
	log            *slog.Logger
}

func NewClient(apiKey string, log *slog.Logger) *Client {
	return &Client{
		apiKey:         apiKey,
		completionAddr: "https://api.openai.com/v1/completions",
		log:            log,
	}
}

// Summarize uses the openAI api to perform a summarization of the given text.
func (c *Client) Summarize(ctx context.Context, text string) (string, error) {
	c.log.Info("summarize text with openAI",
		"method", "Summarize",
		"lenText", len(text))

	if text == "" {
		return "", errors.New("could not summarize, text is empty")
	}

	dto := completionDTO{
		Model:       gpt3TextModel,
		Prompt:      fmt.Sprintf("%s\n\n%s", text, sumPrompt),
		Temperature: 1,
		MaxTokens:   220,
		TopP:        1.0,
		FrequencyP:  0.0,
		PresenceP:   1,
	}

	return c.process(ctx, dto)
}

// NER uses the openAI api to perform named entity recognition of the given text.
// The returned entity types are person, location and organisation.
func (c *Client) NER(ctx context.Context, text string) (article.NER, error) {
	c.log.Info("perform named entity recognition with openAI",
		"method", "NER",
		"lenText", len(text))

	if text == "" {
		return article.NER{}, errors.New("could not perform ner, text is empty")
	}

	dto := completionDTO{
		Model:       gpt3TextModel,
		Prompt:      fmt.Sprintf("%s:\n%s", nerPrompt, text),
		Temperature: 1,
		MaxTokens:   220,
		TopP:        1.0,
		FrequencyP:  0.0,
		PresenceP:   1,
	}

	result, err := c.process(ctx, dto)
	if err != nil {
		return article.NER{}, err
	}

	return c.toNER(result)
}

// process processes the request to openAI and returns the response as text.
// The given completionDTO is encoded and posted to the openAI api. The response is
// unpacked and the content is returned as text.
func (c *Client) process(ctx context.Context, dto completionDTO) (string, error) {
	c.log.Debug("process openAI request", "method", "process")

	payload, err := encoding.EncodeToReader(dto)
	if err != nil {
		return "", err
	}

	var response responseDTO
	err = request.Post(ctx, c.completionAddr, c.getHeader(), payload, &response)
	if err != nil {
		return "", err
	}
	c.log.Debug("response dto", "method", "process", "response", response)

	result, err := c.resultText(response)
	if err != nil {
		return "", err
	}

	return result, nil
}

// getHeader returns a map with Content-Type and Authorization set for openAI.
func (c *Client) getHeader() map[string][]string {
	return map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", c.apiKey)},
	}
}

// resultText extracts the result text from the response and returns
// it as a string. In case multipe results are present it picks the first.
// If no result, or an empty result is found, an ErrInvalidResponse is returned.
func (c *Client) resultText(response responseDTO) (string, error) {
	c.log.Debug("get result text from response", "method", "resultText")

	choices := response.Choices
	if len(choices) == 0 {
		return "", ErrInvalidResponse
	}

	// In case of multiple results, we simply return the first.
	// If in the future multiple results should be used, resultText
	// should return a slice of results.
	if len(choices) > 1 {
		c.log.Info("openAI returned multiple results, use first result",
			"method", "resultText",
			"resultLen", len(choices))
	}

	text := response.Choices[0].Text
	if text == "" {
		return "", ErrInvalidResult
	}

	return text, nil
}

// toNER transforms the given text into an article.NER. The text is
// expected to be the string respresentation of a JSON with can be unmarshalled
// into an nerDTO.
func (c *Client) toNER(text string) (article.NER, error) {
	c.log.Debug("get namedEntities from result text", "method", "toNER")

	if text == "" {
		return article.NER{}, errors.New("text is empty")
	}

	var ner nerDTO
	err := encoding.UnmashalJSON(text, &ner)
	if err != nil {
		return article.NER{}, err
	}

	c.log.Debug("check unmarshal result text", "method", "toNER",
		"Persons", ner.Person,
		"Locations", ner.Location,
		"Organisations", ner.Organisation)

	return article.NER{
		Pers: ner.Person,
		Locs: ner.Location,
		Orgs: ner.Organisation,
	}, nil
}
