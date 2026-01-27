package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new logger based on the provided configuration
func New(level, format, output string) (*zap.Logger, error) {
	// Parse log level
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", level, err)
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// Create encoder based on format
	var encoder zapcore.Encoder
	if format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create writer based on output
	var writer zapcore.WriteSyncer
	if output == "stdout" {
		writer = zapcore.AddSync(os.Stdout)
	} else if output == "stderr" {
		writer = zapcore.AddSync(os.Stderr)
	} else {
		// Output to file
		file, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %q: %w", output, err)
		}
		writer = zapcore.AddSync(file)
	}

	// Create core
	core := zapcore.NewCore(encoder, writer, zapLevel)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}
