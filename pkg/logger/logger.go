/**
 * 功能：logger.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

// Package logger documentation
package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// InitLogger initializes the zap logger
func InitLogger(levelStr string, isDev bool, filename string) error {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(levelStr)); err != nil {
		level = zapcore.InfoLevel
	}

	var config zap.Config
	if isDev {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	config.Level = zap.NewAtomicLevelAt(level)

	// 默认输出到 stdout；如配置了文件路径，则同时输出到文件
	outputPaths := []string{"stdout"}
	if filename != "" {
		logDir := filepath.Dir(filename)
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			return err
		}
		outputPaths = append(outputPaths, filename)
	}
	config.OutputPaths = outputPaths
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	Log = logger
	zap.ReplaceGlobals(logger)

	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
