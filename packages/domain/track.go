package domain

// Track is a song entry within a Setlist. ArtistName is denormalized: a track can
// appear in a merged festival playlist and must carry its own attribution for the
// Spotify search query (track:{title} artist:{artistName}).
type Track struct {
	Title      string
	ArtistName string
}
