package logger

import (
	"fmt"
	"log"

	"go.uber.org/zap"

	"go.uber.org/zap/zapcore"
)

// LoggingConfig defines the necessary configuration for instantiating a Logger
type LoggingConfig struct {
	Level            string
	AppVersion       string
	GitSha           string
	OutputPaths      []string
	ErrorOutputPaths []string
}

// Logger is a zap logger. If performance is a concern, use this logger.
var Logger = zap.NewNop()

// SugaredLogger abstracts away types and lets the zap library figure
// them out so that the caller doesn't have to import zap into their package
// but is slightly slower and creates more garbage.
var SugaredLogger = Logger.Sugar()

// InitializeLogger sets up the logger. This function should be called as soon
// as possible. Any use of the logger provided by this package will be a nop
// until this function is called
func (lc *LoggingConfig) InitializeLogger() {
	var err error
	var logConfig zap.Config
	var level zapcore.Level
	if err := level.Set(lc.Level); err != nil {
		fmt.Printf("Invalid log level %s. Using INFO.", lc.Level)
		level.Set("info")
	}
	logConfig = zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       false,
		DisableStacktrace: false,
		Encoding:          "json",
		EncoderConfig:     zap.NewProductionEncoderConfig(),
		OutputPaths:       append(lc.OutputPaths, "stdout"),
		ErrorOutputPaths:  append(lc.ErrorOutputPaths, "stderr"),
		InitialFields:     map[string]interface{}{"appVersion": lc.AppVersion, "gitSha": lc.GitSha},
	}
	Logger, err = logConfig.Build()
	if err != nil {
		fmt.Printf("Error initializing Logger: %s\n", err.Error())
	} else {
		Logger.Info("Logger initialized")
	}
	SugaredLogger = Logger.Sugar()
}

// CreateStdLogger returns a standard-library compatible logger
func CreateStdLogger(zapLogger *zap.Logger, logLevel string) (*log.Logger, error) {
	switch {
	case logLevel == "debug":
		return zap.NewStdLogAt(zapLogger, zapcore.DebugLevel)
	case logLevel == "info":
		return zap.NewStdLogAt(zapLogger, zapcore.InfoLevel)
	case logLevel == "warn":
		return zap.NewStdLogAt(zapLogger, zapcore.WarnLevel)
	case logLevel == "error":
		return zap.NewStdLogAt(zapLogger, zapcore.ErrorLevel)
	case logLevel == "panic":
		return zap.NewStdLogAt(zapLogger, zapcore.PanicLevel)
	case logLevel == "fatal":
		return zap.NewStdLogAt(zapLogger, zapcore.FatalLevel)
	}
	return nil, fmt.Errorf("Unknown log level %s", logLevel)
}
