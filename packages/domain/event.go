package domain

import "time"

// Event represents a concert or festival. Artists preserves source order — this order
// determines track grouping in merged festival playlists.
type Event struct {
	Name     string
	Date     time.Time
	Location string
	Artists  []Artist // lineup in source order; order must be preserved through all layers
}
