// Package observability wires third-party observability SDKs (Sentry) to the
// usecases.ErrorReporter port. The usecases package has no Sentry dependency;
// composition happens here at the app boundary.
package observability

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

// SentryReporter implements usecases.ErrorReporter by forwarding errors to Sentry.
type SentryReporter struct{}

// Capture sends the error to Sentry. Safe to call with a nil error (no-op).
func (SentryReporter) Capture(err error) {
	if err == nil {
		return
	}
	sentry.CaptureException(err)
}

// InitSentry initialises the Sentry SDK. Returns a flush function that must be
// deferred by the caller. When dsn is empty, Sentry is not initialised and the
// flush function is a no-op — the reporter then becomes effectively silent.
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
