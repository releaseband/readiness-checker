package readiness

import (
	"fmt"
	"log/slog"

	golog "github.com/InVisionApp/go-logger"
)

var _ golog.Logger = (*slogAdapter)(nil) // compile-time interface check

// slogAdapter bridges *slog.Logger to go-health's internal Logger interface
// (github.com/InVisionApp/go-logger).
type slogAdapter struct{ l *slog.Logger }

func NewSlogAdapter(logger *slog.Logger) *slogAdapter {
	return &slogAdapter{l: logger}
}

func (a *slogAdapter) Debug(msg ...any) { a.l.Debug(fmt.Sprint(msg...)) }
func (a *slogAdapter) Info(msg ...any)  { a.l.Info(fmt.Sprint(msg...)) }
func (a *slogAdapter) Warn(msg ...any)  { a.l.Warn(fmt.Sprint(msg...)) }
func (a *slogAdapter) Error(msg ...any) { a.l.Error(fmt.Sprint(msg...)) }

func (a *slogAdapter) Debugln(msg ...any) { a.l.Debug(fmt.Sprintln(msg...)) }
func (a *slogAdapter) Infoln(msg ...any)  { a.l.Info(fmt.Sprintln(msg...)) }
func (a *slogAdapter) Warnln(msg ...any)  { a.l.Warn(fmt.Sprintln(msg...)) }
func (a *slogAdapter) Errorln(msg ...any) { a.l.Error(fmt.Sprintln(msg...)) }

func (a *slogAdapter) Debugf(format string, args ...any) {
	a.l.Debug(fmt.Sprintf(format, args...))
}
func (a *slogAdapter) Infof(format string, args ...any) {
	a.l.Info(fmt.Sprintf(format, args...))
}
func (a *slogAdapter) Warnf(format string, args ...any) {
	a.l.Warn(fmt.Sprintf(format, args...))
}
func (a *slogAdapter) Errorf(format string, args ...any) {
	a.l.Error(fmt.Sprintf(format, args...))
}

func (a *slogAdapter) WithFields(fields golog.Fields) golog.Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &slogAdapter{a.l.With(args...)}
}
