module gigtape/cli

go 1.22

require (
	gigtape/adapters/setlistfm v0.0.0
	gigtape/adapters/spotify v0.0.0
	gigtape/domain v0.0.0
	gigtape/usecases v0.0.0
	github.com/spf13/cobra v1.8.1
	golang.org/x/oauth2 v0.20.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace (
	gigtape/adapters/setlistfm => ../../packages/adapters/setlistfm
	gigtape/adapters/spotify => ../../packages/adapters/spotify
	gigtape/domain => ../../packages/domain
	gigtape/usecases => ../../packages/usecases
)
