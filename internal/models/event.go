package models

import (
	"batch-saver/api"
	"errors"
)

type Event struct {
	ID      string
	GroupID string
	Data    []byte
}

func EventFromAPI(in *api.Event) (Event, error) {
	if in == nil || in.Id == "" || in.GroupId == "" || in.Data == nil {
		return Event{}, errors.New("invalid event")
	}

	return Event{
		ID:      in.Id,
		GroupID: in.GroupId,
		Data:    in.Data,
	}, nil
}
