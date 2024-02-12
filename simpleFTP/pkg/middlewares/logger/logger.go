package logger

import (
	"log"

	"go.uber.org/zap"
)

var logger *zap.Logger
var level zap.AtomicLevel

func init() {
	cfg := zap.NewProductionConfig()
	level = cfg.Level
	l, err := cfg.Build()
	if err != nil {
		log.Fatalf("unable to initialize logger, error %v.", err.Error())
	}
	logger = l
}

func Logger() *zap.Logger {
	return logger
}

func SugaredLogger() *zap.SugaredLogger {
	return logger.Sugar()
}

func Level() zap.AtomicLevel {
	return level
}

func UseDevelopmentLogger() {
	cfg := zap.NewDevelopmentConfig()
	level = cfg.Level
	l, err := cfg.Build()
	if err != nil {
		log.Fatalf("unable to initiliaze logger, error %v", err.Error())
	}
	logger = l
}
