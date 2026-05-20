package logger

import (
	"go-test/internal/global"

	"go.uber.org/zap"
)

func InitLogger() error {
	l, err := zap.NewProduction()
	if err != nil {
		return err
	}

	global.Logger = l
	return nil
}