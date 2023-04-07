package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(devLogger bool) (*zap.SugaredLogger, error) {
	if devLogger {
		return NewDevelopment()
	}

	return NewProduction()
}

func NewDevelopment() (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.InitialFields = map[string]any{"service": "articleDB"}
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	cfg.Encoding = "console"
	cfg.DisableStacktrace = true

	zapper, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return zapper.Sugar(), err
}

func NewProduction() (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.InitialFields = map[string]any{"service": "articleDB"}
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	cfg.Encoding = "console"
	cfg.DisableStacktrace = true

	zapper, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return zapper.Sugar(), err
}

func NewTest(debug bool) (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.InitialFields = map[string]any{"service": "articleDB"}
	cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	cfg.Encoding = "console"

	if debug {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	zapper, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return zapper.Sugar(), err
}
