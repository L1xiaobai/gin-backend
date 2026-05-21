package logger

import (
	"go-test/internal/global"

	"go.uber.org/zap"
)

func InitLogger() error {
	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = true

	l, err := cfg.Build()
	if err != nil {
		return err
	}

	global.Logger = l
	return nil
}