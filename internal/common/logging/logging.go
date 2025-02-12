package logging

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
}

func CreateLogger() (*zap.SugaredLogger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("error building logger: %w", err)
	}

	return logger.Sugar(), nil
}

type VoidLogger struct{}

func NewVoidLogger() *VoidLogger {
	return &VoidLogger{}
}

func (l VoidLogger) Infow(msg string, keysAndValues ...interface{})  {} //nolint:revive
func (l VoidLogger) Errorw(msg string, keysAndValues ...interface{}) {} //nolint:revive
func (l VoidLogger) Warnw(msg string, keysAndValues ...interface{})  {} //nolint:revive
func (l VoidLogger) Debugw(msg string, keysAndValues ...interface{}) {} //nolint:revive
