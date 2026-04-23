package fakes

import (
	"context"

	"gigtape/domain"
)

// FakeEventProvider implements domain.EventProvider with configurable return values.
type FakeEventProvider struct {
	Events []domain.Event
	Err    error

	SearchEventsCalledWith string
}

func (f *FakeEventProvider) SearchEvents(_ context.Context, name string) ([]domain.Event, error) {
	f.SearchEventsCalledWith = name
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Events, nil
}
