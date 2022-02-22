package core

import (
	"bytes"
	"encoding/json"
)

const (
	ButtonLeft       int = 1
	ButtonMiddle         = 3
	ButtonRight          = 3
	ButtonScrollUp       = 4
	ButtonScrollDown     = 5
)

type I3barHeader struct {
	Version     int  `json:"version"`
	ClickEvents bool `json:"click_events,omitempty"`
	ContSignal  int  `json:"cont_signal,omitempty"`
	StopSignal  int  `json:"stop_signal,omitempty"`
}

type I3barClickEvent struct {
	Name      string `json:"name"`
	Instance  string `json:"instance"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Button    int    `json:"button"`
	Event     int    `json:"event"`
	RelativeX int    `json:"relative_x"`
	RelativeY int    `json:"relative_y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

func NewEventFromRaw(raw []byte) (*I3barClickEvent, error) {
	raw = bytes.TrimLeftFunc(raw, func(r rune) bool {
		return r != '{'
	})
	raw = bytes.TrimRightFunc(raw, func(r rune) bool {
		return r != '}'
	})
	ev := new(I3barClickEvent)
	if err := json.Unmarshal(raw, ev); err != nil {
		return nil, err
	}
	return ev, nil
}

type I3barBlock struct {
	FullText   string `json:"full_text"`
	ShortText  string `json:"short_text,omitempty"`
	Color      string `json:"color,omitempty"`
	Background string `json:"background,omitempty"`
	Border     string `json:"border,omitempty"`
	BorderTop  string `json:"border_top,omitempty"`
	// Omitted...
	Name string `json:"name,omitempty"`
	// Omitted...
	Markdup string `json:"markup,omitempty"`
}
