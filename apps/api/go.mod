module gigtape/api

go 1.22

require (
	gigtape/adapters/setlistfm v0.0.0
	gigtape/adapters/spotify v0.0.0
	gigtape/domain v0.0.0
	gigtape/usecases v0.0.0
	github.com/getsentry/sentry-go v0.27.0
	github.com/gin-gonic/gin v1.10.0
	golang.org/x/oauth2 v0.20.0
)

replace (
	gigtape/adapters/setlistfm => ../../packages/adapters/setlistfm
	gigtape/adapters/spotify => ../../packages/adapters/spotify
	gigtape/domain => ../../packages/domain
	gigtape/usecases => ../../packages/usecases
)
