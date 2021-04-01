package adapter

import (
	"fmt"
	"github.com/InVisionApp/go-logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapAdapter struct {
	logger *zap.Logger
	fields map[string]interface{}
}

func NewZapLoggerAdapter(logger *zap.Logger) *ZapAdapter {
	return &ZapAdapter{
		logger: logger,
	}
}

func parseString(msg ...interface{}) string {
	var str string
	for _, m := range msg {
		s, ok := m.(string)
		if ok {
			str += s
		} else {
			str += fmt.Sprint(s)
		}
	}

	return str
}

func (z *ZapAdapter) pretty() []zap.Field {
	fields := make([]zap.Field, len(z.fields))

	i := 0
	for k, v := range z.fields {
		fields[i] = zap.Any(k, v)
		i++
	}

	z.fields = nil

	return fields
}

func (z *ZapAdapter) log(level zapcore.Level, msg ...interface{}) {
	logger := z.logger

	if len(z.fields) > 0 {
		logger = logger.With(z.pretty()...)
	}

	logger.Check(level, parseString(msg...))
}

func (z *ZapAdapter) Logf(level zapcore.Level, format string, args ...interface{}) {
	logger := z.logger

	if len(z.fields) > 0 {
		logger = logger.With(z.pretty()...)
	}

	logger.Check(level, fmt.Sprintf(format, args...))
}

func (z *ZapAdapter) Debug(msg ...interface{}) {
	z.log(zap.DebugLevel, msg...)
}

func (z *ZapAdapter) Info(msg ...interface{}) {
	z.log(zap.InfoLevel, msg...)
}

func (z *ZapAdapter) Warn(msg ...interface{}) {
	z.log(zap.WarnLevel, msg...)
}

func (z *ZapAdapter) Error(msg ...interface{}) {
	z.log(zap.ErrorLevel, msg...)
}

func (z *ZapAdapter) Debugln(msg ...interface{}) {
	z.log(zap.DebugLevel, msg...)
}

func (z *ZapAdapter) Infoln(msg ...interface{}) {
	z.log(zap.InfoLevel, msg...)
}

func (z *ZapAdapter) Warnln(msg ...interface{}) {
	z.log(zap.WarnLevel, msg...)
}

func (z *ZapAdapter) Errorln(msg ...interface{}) {
	z.log(zap.ErrorLevel, msg...)
}

func (z *ZapAdapter) Debugf(format string, args ...interface{}) {
	z.Logf(zap.DebugLevel, format, args...)
}

func (z *ZapAdapter) Infof(format string, args ...interface{}) {
	z.Logf(zap.InfoLevel, format, args...)
}

func (z *ZapAdapter) Warnf(format string, args ...interface{}) {
	z.Logf(zap.WarnLevel, format, args...)
}

func (z *ZapAdapter) Errorf(format string, args ...interface{}) {
	z.Logf(zap.ErrorLevel, format, args...)
}

func (z *ZapAdapter) WithFields(fields log.Fields) log.Logger {
	cp := &ZapAdapter{
		logger: z.logger,
	}

	if z.fields == nil {
		cp.fields = fields
		return cp
	}

	cp.fields = make(map[string]interface{}, len(z.fields)+len(fields))
	for k, v := range z.fields {
		cp.fields[k] = v
	}

	for k, v := range fields {
		cp.fields[k] = v
	}

	return cp
}
