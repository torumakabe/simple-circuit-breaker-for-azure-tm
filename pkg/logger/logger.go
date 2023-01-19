package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var zapLog *zap.SugaredLogger

var logLevelSeverity = map[zapcore.Level]string{
	zapcore.DebugLevel:  "DEBUG",
	zapcore.InfoLevel:   "INFO",
	zapcore.WarnLevel:   "WARNING",
	zapcore.ErrorLevel:  "ERROR",
	zapcore.DPanicLevel: "CRITICAL",
	zapcore.PanicLevel:  "CRITICAL",
	zapcore.FatalLevel:  "CRITICAL",
}

func init() {
	highPriority := zap.LevelEnablerFunc(func(lv zapcore.Level) bool {
		return lv >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})
	stdoutSink := zapcore.Lock(os.Stdout)
	stderrSink := zapcore.Lock(os.Stderr)

	enc := zapcore.NewConsoleEncoder(newEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(enc, stderrSink, highPriority),
		zapcore.NewCore(enc, stdoutSink, lowPriority),
	)

	l := zap.New(core, zap.AddCallerSkip(1))
	zapLog = l.Sugar()
}

func EncodeLevel(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(logLevelSeverity[l])
}

func newEncoderConfig() zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeLevel = EncodeLevel
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder

	return cfg
}

func Debug(args ...interface{}) {
	zapLog.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	zapLog.Debugf(template, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	zapLog.Debugw(msg, keysAndValues...)
}

func Info(args ...interface{}) {
	zapLog.Info(args...)
}

func Infof(template string, args ...interface{}) {
	zapLog.Infof(template, args...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	zapLog.Infow(msg, keysAndValues...)
}

func Error(args ...interface{}) {
	zapLog.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	zapLog.Errorf(template, args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	zapLog.Errorw(msg, keysAndValues...)
}

func Fatal(args ...interface{}) {
	zapLog.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	zapLog.Fatalf(template, args...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	zapLog.Fatalw(msg, keysAndValues...)
}

func Sync() error {
	return zapLog.Sync()
}
