package log

import (
	"os"
	"time"

	"github.com/chiehting/gitlab-record-collection/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// error logger
var _errorLogger *zap.SugaredLogger

var _levelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func getLoggerLevel(lvl string) zapcore.Level {
	if level, ok := _levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}

func init() {
	cfg := config.GetLog()
	filePath := cfg.FilePath
	level := getLoggerLevel(cfg.Level)

	hook := lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    128, // megabytes
		MaxBackups: 100, // backup count
		MaxAge:     30,  // days
		Compress:   true,
		LocalTime:  true,
	}

	syncWriter := zapcore.AddSync(&hook)

	encoder := zap.NewProductionEncoderConfig()
	if cfg.OmitTimeKey {
		encoder.TimeKey = zapcore.OmitKey
	}
	encoder.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.UTC().Format("2006-01-02T15:04:05Z0700"))
	})
	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleEncoder := zapcore.NewConsoleEncoder(encoder)
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoder), syncWriter, zap.NewAtomicLevelAt(level)),
		zapcore.NewCore(consoleEncoder, consoleDebugging, zap.NewAtomicLevelAt(level)),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	_errorLogger = logger.Sugar()

	Infof("log.level:%s", level)
}

// Debug show debug message
func Debug(args ...interface{}) {
	_errorLogger.Debug(args...)
}

// Debugf show debug message by format
func Debugf(template string, args ...interface{}) {
	_errorLogger.Debugf(template, args...)
}

// Info show debug message
func Info(args ...interface{}) {
	_errorLogger.Info(args...)
}

// Infof show debug message by format
func Infof(template string, args ...interface{}) {
	_errorLogger.Infof(template, args...)
}

// Warn show debug message
func Warn(args ...interface{}) {
	_errorLogger.Warn(args...)
}

// Warnf show debug message by format
func Warnf(template string, args ...interface{}) {
	_errorLogger.Warnf(template, args...)
}

// Error show debug message
func Error(args ...interface{}) {
	_errorLogger.Error(args...)
}

// Errorf show debug message by format
func Errorf(template string, args ...interface{}) {
	_errorLogger.Errorf(template, args...)
}

// DPanic show debug message
func DPanic(args ...interface{}) {
	_errorLogger.DPanic(args...)
}

// DPanicf show debug message by format
func DPanicf(template string, args ...interface{}) {
	_errorLogger.DPanicf(template, args...)
}

// Panic show debug message
func Panic(args ...interface{}) {
	_errorLogger.Panic(args...)
}

// Panicf show debug message by format
func Panicf(template string, args ...interface{}) {
	_errorLogger.Panicf(template, args...)
}

// Fatal show debug message
func Fatal(args ...interface{}) {
	_errorLogger.Fatal(args...)
}

// Fatalf show debug message by format
func Fatalf(template string, args ...interface{}) {
	_errorLogger.Fatalf(template, args...)
}
