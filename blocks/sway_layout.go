package blocks

import (
	"context"
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
	go func() {
		<-ctx.Done()
		ipcClient.Close()
	}()
	s.ipc = ipcClient
	var err error
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
	if _, res, err = ipcClient.Recv(); err != nil {
		return
	}
	now := time.Now()
	for {
		t, res, err := ipcClient.Recv()
		if err != nil {
			return
		}
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
			"flag":  CountryFlagFromIsoCode(shortName),
		}),
	}}
}

func init() {
	RegisterBlocklet("sway_layout", NewSwayLayoutBlock)
}
