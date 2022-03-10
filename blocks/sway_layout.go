package blocks

import (
	"context"
	"strings"
	"time"

	"github.com/kraftwerk28/gost/blocks/ipc"
	"github.com/kraftwerk28/gost/blocks/rxkbcommon"
	. "github.com/kraftwerk28/gost/core"
	"github.com/kraftwerk28/gost/core/formatting"
)

type SwayLayoutConfig struct {
	Format *ConfigFormat `yaml:"format"`
}

type SwayLayout struct {
	SwayLayoutConfig
	layoutLongToShort  map[string]string
	layouts            []string
	currentLayoutIndex int
	ipc                *ipc.IpcClient
}

func NewSwayLayoutBlock() I3barBlocklet {
	return &SwayLayout{layoutLongToShort: rxkbcommon.GetLayoutShortNames()}
}

func (s *SwayLayout) GetConfig() interface{} {
	return &s.SwayLayoutConfig
}

func (s *SwayLayout) Run(ch UpdateChan, ctx context.Context) {
	ipcClient, _ := ipc.NewIpcClient()
	s.ipc = ipcClient
	ipcClient.SendRaw(ipc.IpcMsgTypeGetInputs, nil)
	_, res, _ := ipcClient.Recv()
	for _, device := range *res.(*[]ipc.IpcInputDevice) {
		if device.Type == "keyboard" {
			s.currentLayoutIndex = device.XkbActiveLayoutIndex
			s.layouts = device.XkbLayoutNames
			ch.SendUpdate()
			break
		}
	}
	ipcClient.SendRaw(ipc.IpcMsgTypeSubscribe, []byte(`["input"]`))
	_, res, _ = ipcClient.Recv()
	now := time.Now()
	for {
		t, res, _ := ipcClient.Recv()
		if t != ipc.IpcEventTypeInput {
			continue
		}
		device := res.(*ipc.InputChange).Input
		if device.Type != "keyboard" {
			continue
		}
		s.currentLayoutIndex = device.XkbActiveLayoutIndex
		s.layouts = device.XkbLayoutNames
		now2 := time.Now()
		if now2.Sub(now).Milliseconds() < 1 {
			continue
		}
		now = now2
		ch.SendUpdate()
	}
}

func countryFlagFromIsoCode(countryCode string) string {
	if len(countryCode) != 2 {
		return countryCode
	}
	b := []byte(strings.ToUpper(countryCode))
	// Each char is encoded as 1F1E6 to 1F1FF for A-Z
	c1, c2 := b[0]+0xa5, b[1]+0xa5
	// The last byte will always start with 101 (0xa0) and then the 5 least
	// significant bits from the previous result
	b1 := 0xa0 | (c1 & 0x1f)
	b2 := 0xa0 | (c2 & 0x1f)
	// Get the flag string from the UTF-8 representation of our Unicode characters.
	return string([]byte{0xf0, 0x9f, 0x87, b1, 0xf0, 0x9f, 0x87, b2})
}

func (s *SwayLayout) OnEvent(e *I3barClickEvent, ctx context.Context) {
	if e.Button == ButtonRight {
		s.ipc.SendRaw(
			ipc.IpcMsgTypeCommand,
			[]byte(`input * xkb_switch_layout next`),
		)
	}
}

func (s *SwayLayout) Render() []I3barBlock {
	if s.layouts == nil {
		return nil
	}
	currentLayout := s.layouts[s.currentLayoutIndex]
	shortName := s.layoutLongToShort[currentLayout]
	return []I3barBlock{{
		FullText: s.Format.Expand(formatting.NamedArgs{
			"long":  currentLayout,
			"short": shortName,
			"flag":  countryFlagFromIsoCode(shortName),
		}),
	}}
}

func init() {
	RegisterBlocklet("sway_layout", NewSwayLayoutBlock)
}
