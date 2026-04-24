module gigtape/adapters/spotify

go 1.22

require (
	gigtape/domain v0.0.0
	golang.org/x/oauth2 v0.20.0
)

require github.com/stretchr/testify v1.9.0 // indirect

replace gigtape/domain => ../../domain
