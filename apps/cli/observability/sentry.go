// Package observability wires Sentry to the usecases.ErrorReporter port for
// the CLI. Mirrors apps/api/observability.
package observability

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

// SentryReporter implements usecases.ErrorReporter.
type SentryReporter struct{}

// Capture forwards the error to Sentry.
func (SentryReporter) Capture(err error) {
	if err == nil {
		return
	}
	sentry.CaptureException(err)
}

// InitSentry initialises Sentry. Returns a flush function the caller must
// defer. When dsn is empty, Sentry is left uninitialised.
func InitSentry(dsn, environment, release string) (flush func(), err error) {
	if dsn == "" {
		return func() {}, nil
	}
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Environment: environment,
		Release:     release,
	}); err != nil {
		return func() {}, fmt.Errorf("sentry init: %w", err)
	}
	return func() { sentry.Flush(2 * time.Second) }, nil
}
