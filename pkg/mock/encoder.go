package mock

import (
	"context"

	"github.com/Br0ce/articleDB/pkg/vector"
)

type Encoder struct {
	EncodeFn      func(ctx context.Context, texts []string) ([]vector.Vector, error)
	EncodeInvoked bool
}

func (e *Encoder) Encode(ctx context.Context, texts []string) ([]vector.Vector, error) {
	e.EncodeInvoked = true
	return e.EncodeFn(ctx, texts)
}
