package domain

// PlaylistResult is the structured outcome of a playlist creation operation.
// It is always returned non-nil by use cases — never replaced with a plain error.
// A result with a non-empty PlaylistURL and non-empty UnmatchedTracks is a valid
// partial success: the playlist was created, some tracks were not found.
type PlaylistResult struct {
	PlaylistURL     string   // direct link to the created playlist
	MatchedTracks   []Track  // tracks successfully found and added
	UnmatchedTracks []string // track titles not found in the music service
	SkippedArtists  []string // artists with no setlist and no manual tracks, or deselected
}
