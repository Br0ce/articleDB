package vector

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/Br0ce/articleDB/pkg/encoding"
	"github.com/Br0ce/articleDB/pkg/request"
	"github.com/Br0ce/articleDB/pkg/vector"
)

// Encoder is a client to a remote inference server serving an text embedding model
// abiding by the Predict Protocol, version 2 [kserve].
//
// [kserve]: https://github.com/kserve/kserve/tree/master/docs/predict-api/v2
type Encoder struct {
	inferAddr string
	readyAddr string
	log       *slog.Logger
}

type embedRequest struct {
	Inputs []container[string] `json:"inputs"`
}

type embedResponse struct {
	ModelName    string               `json:"model_name"`
	ModelVersion string               `json:"model_version"`
	Outputs      []container[float32] `json:"outputs"`
}

type container[T any] struct {
	Name     string `json:"name"`
	Shape    []int  `json:"shape"`
	DataType string `json:"datatype"`
	Data     []T    `json:"data"`
}

// NewEncoder returns a new encoder, a local client to a remote model for text embeddings.
func NewEncoder(addr string, modelName string, log *slog.Logger) (*Encoder, error) {
	readyAddr, err := url.JoinPath(addr, fmt.Sprintf("models/%s/ready", modelName))
	if err != nil {
		return nil, fmt.Errorf("cannot join ready path, %s", err.Error())
	}

	inferAddr, err := url.JoinPath(addr, fmt.Sprintf("models/%s/infer", modelName))
	if err != nil {
		return nil, fmt.Errorf("cannot join infer path, %s", err.Error())
	}

	return &Encoder{
		readyAddr: readyAddr,
		inferAddr: inferAddr,
		log:       log,
	}, nil
}

// Encode returns text embeddings for each text in texts in the order given by texts.
func (enc *Encoder) Encode(ctx context.Context, texts []string) ([]vector.Vector, error) {
	enc.log.Info("encode texts", "method", "Encode", "textsLen", len(texts))

	// Shortcut since texts is empty.
	if len(texts) == 0 {
		return []vector.Vector{}, nil
	}

	texts = removeInvalidBytes(texts)
	req := parseEmbedRequest(texts)
	r, err := encoding.EncodeToReader(req)
	if err != nil {
		return nil, fmt.Errorf("cannot create payload, %s", err.Error())
	}

	var resp embedResponse
	err = request.Post(ctx, enc.inferAddr, http.Header{}, r, &resp)
	if err != nil {
		return nil, fmt.Errorf("cannot post request, %s", err.Error())
	}

	return parseVectors(resp), nil
}

// Ready returns true, if the remote encoder is ready for inference.
func (enc *Encoder) Ready() bool {
	if err := request.Get(context.Background(), enc.readyAddr, http.Header{}, nil); err != nil {
		return false
	}
	return true
}

// removeInvalidBytes returns a copy of texts with every non utf8 byte removed.
func removeInvalidBytes(texts []string) []string {
	for i, t := range texts {
		texts[i] = strings.ToValidUTF8(t, "")
	}
	return texts
}

// parseVectors maps an embedResponse to vector.Vectors.
func parseVectors(resp embedResponse) []vector.Vector {
	vv := make([]vector.Vector, len(resp.Outputs))
	for i, o := range resp.Outputs {
		vv[i] = vector.Vector{
			ID:   "", // todo
			Data: o.Data,
		}
	}
	return vv
}

// parseEmbedRequest maps a slice of strings an embedRequest.
func parseEmbedRequest(texts []string) embedRequest {
	return embedRequest{
		Inputs: []container[string]{
			{
				Name:     "sentences",
				Shape:    []int{len(texts)},
				DataType: "BYTES",
				Data:     texts,
			},
		},
	}
}
