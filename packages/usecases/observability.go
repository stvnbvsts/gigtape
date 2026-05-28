package usecases

import "log/slog"

// ErrorReporter is the use-case-level port for unexpected error reporting
// (e.g. Sentry). The default implementation is NoopReporter; delivery apps
// wire a concrete adapter at composition time.
type ErrorReporter interface {
	Capture(err error)
}

// NoopReporter drops all errors. Used when no reporter is injected.
type NoopReporter struct{}

func (NoopReporter) Capture(error) {}

// defaultLogger returns lg if non-nil, otherwise slog.Default().
func defaultLogger(lg *slog.Logger) *slog.Logger {
	if lg != nil {
		return lg
	}
	return slog.Default()
}

// defaultReporter returns r if non-nil, otherwise NoopReporter.
func defaultReporter(r ErrorReporter) ErrorReporter {
	if r != nil {
		return r
	}
	return NoopReporter{}
}
