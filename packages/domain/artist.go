// Package domain is the architectural center of Gigtape. It has zero external imports;
// the Go compiler enforces this boundary. Adapters import domain to implement port
// interfaces; domain never imports adapters.
package domain

// Artist represents a musical performer. ExternalRef is an opaque adapter token
// (e.g. a MusicBrainz ID from setlist.fm); the domain never interprets its value.
type Artist struct {
	Name           string
	Disambiguation string // e.g. "rock band from Abingdon, UK"; shown in confirmation step
	ExternalRef    string
}
