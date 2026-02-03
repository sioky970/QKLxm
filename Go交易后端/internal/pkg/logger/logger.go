package logger

import (
	"os"
	"path/filepath"

	"exchange-go/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 全局日志实例
var Logger *zap.SugaredLogger

// Init 初始化日志
func Init(cfg *config.LogConfig) error {
	// 确保日志目录存在
	logDir := filepath.Dir(cfg.Filename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// 日志级别
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 文件写入器 (日志轮转)
	fileWriter := &lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	// 创建Core
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(fileWriter),
		level,
	)

	// 控制台输出
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	// 合并多个Core
	core := zapcore.NewTee(fileCore, consoleCore)

	// 创建Logger
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	Logger = logger.Sugar()

	return nil
}

// Debug 输出Debug级别日志
func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

// Debugf 格式化输出Debug级别日志
func Debugf(template string, args ...interface{}) {
	Logger.Debugf(template, args...)
}

// Info 输出Info级别日志
func Info(args ...interface{}) {
	Logger.Info(args...)
}

// Infof 格式化输出Info级别日志
func Infof(template string, args ...interface{}) {
	Logger.Infof(template, args...)
}

// Warn 输出Warn级别日志
func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

// Warnf 格式化输出Warn级别日志
func Warnf(template string, args ...interface{}) {
	Logger.Warnf(template, args...)
}

// Error 输出Error级别日志
func Error(args ...interface{}) {
	Logger.Error(args...)
}

// Errorf 格式化输出Error级别日志
func Errorf(template string, args ...interface{}) {
	Logger.Errorf(template, args...)
}

// Fatal 输出Fatal级别日志并退出
func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

// Fatalf 格式化输出Fatal级别日志并退出
func Fatalf(template string, args ...interface{}) {
	Logger.Fatalf(template, args...)
}

// Sync 同步日志缓冲
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
