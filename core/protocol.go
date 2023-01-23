package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type I3barMarkup string

const (
	MarkupNone  I3barMarkup = "none"
	MarkupPango I3barMarkup = "pango"
)

type eventButton int

func (b eventButton) String() string {
	switch b {
	case ButtonLeft:
		return "Left"
	case ButtonMiddle:
		return "Middle"
	case ButtonRight:
		return "Right"
	case ButtonScrollUp:
		return "ScrollUp"
	case ButtonScrollDown:
		return "ScrollDown"
	default:
		return ""
	}
}

const (
	ButtonLeft       eventButton = 1
	ButtonMiddle     eventButton = 2
	ButtonRight      eventButton = 3
	ButtonScrollUp   eventButton = 4
	ButtonScrollDown eventButton = 5
)

type I3barHeader struct {
	Version     int  `json:"version"`
	ClickEvents bool `json:"click_events,omitempty"`
	ContSignal  int  `json:"cont_signal,omitempty"`
	StopSignal  int  `json:"stop_signal,omitempty"`
}

type I3barClickEvent struct {
	Name      string      `json:"name"`
	Instance  string      `json:"instance"`
	X         int         `json:"x"`
	Y         int         `json:"y"`
	Button    eventButton `json:"button"`
	Event     int         `json:"event"`
	RelativeX int         `json:"relative_x"`
	RelativeY int         `json:"relative_y"`
	Width     int         `json:"width"`
	Height    int         `json:"height"`
}

func (e *I3barClickEvent) CustomBlockletName() string {
	i := strings.LastIndexByte(e.Name, ':')
	return e.Name[i+1:]
}

func (e *I3barClickEvent) ShellCommand(
	command string,
	ctx context.Context,
) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(
		cmd.Env,
		fmt.Sprintf("BUTTON=%s", e.Button.String()),
		fmt.Sprintf("X=%d", e.X),
		fmt.Sprintf("Y=%d", e.Y),
	)
	cmd.Stdout = Log.Writer()
	cmd.Stderr = Log.Writer()
	return cmd
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

type blockAlign string

const (
	BlockAlignLeft   blockAlign = "left"
	BlockAlignRight  blockAlign = "right"
	BlockAlignCenter blockAlign = "center"
)

type I3barBlock struct {
	// The text that will be displayed. If missing, the block will be skipped.
	FullText string `json:"full_text"`

	// If given and the text needs to be shortened due to space, this will be
	// displayed instead of full_text
	ShortText string `json:"short_text,omitempty"`

	// The text color to use in #RRGGBBAA or #RRGGBB notation
	Color string `json:"color,omitempty"`

	// The background color for the block in #RRGGBBAA or #RRGGBB notation
	Background string `json:"background,omitempty"`

	// The border color for the block in #RRGGBBAA or #RRGGBB notation
	Border string `json:"border,omitempty"`

	// The height in pixels of the top border. The default is 1
	BorderTop int `json:"border_top,omitempty"`

	// The height in pixels of the bottom border. The default is 1
	BorderBottom int `json:"border_bottom,omitempty"`

	// The width in pixels of the left border. The default is 1
	BorderLeft int `json:"border_left,omitempty"`

	// The width in pixels of the right border. The default is 1
	BorderRight int `json:"border_right,omitempty"`

	// TODO: custom type
	MinWidth int `json:"min_width"`

	// If the text does not span the full width of the block, this specifies
	// how the text should be aligned inside of the block. This can be left
	// (default), right, or center.
	Align blockAlign `json:"align,omitempty"`

	// A name for the block. This is only used to identify the block for click
	// events. If set, each block should have a unique name and instance pair.
	Name string `json:"name,omitempty"`

	// The instance of the name for the block. This is only used to identify the
	// block for click events. If set, each block should have a unique name and
	// instance pair.
	Instance string `json:"instance,omitempty"`

	// Whether the block should be displayed as urgent. Currently swaybar
	// utilizes the colors set in the sway config for urgent workspace buttons.
	// See sway-bar(5) for more information on bar color configuration.
	Urgent bool `json:"urgent,omitempty"`

	// Whether the bar separator should be drawn after the block. See
	// sway-bar(5) for more information on how to set the separator text.
	Separator bool `json:"separator,omitempty"`

	// The amount of pixels to leave blank after the block. The separator text
	// will be displayed centered in this gap. The default is 9 pixels.
	SeparatorBlockWidth int `json:"separator_block_width,omitempty"`

	// The type of markup to use when parsing the text for the block. This can
	// either be pango or none (default).
	Markup I3barMarkup `json:"markup,omitempty"`
}
