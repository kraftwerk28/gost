package blocks

import (
	"context"
	"fmt"
	"time"

	"github.com/kraftwerk28/gost/blocks/ipc"
	"github.com/kraftwerk28/gost/blocks/rxkbcommon"
	. "github.com/kraftwerk28/gost/core"
	"github.com/kraftwerk28/gost/core/formatting"
)

type SwayLayoutConfig struct {
	Format *ConfigFormat `yaml:"format"`
	Input  *string       `yaml:"input"`
}

type SwayLayout struct {
	SwayLayoutConfig
	layoutLongToShort  map[string]string
	layouts            []string
	currentLayoutIndex int
	ipc                *ipc.IpcClient
}

func NewSwayLayoutBlock() I3barBlocklet {
	return &SwayLayout{}
}

func (s *SwayLayout) GetConfig() interface{} {
	return &s.SwayLayoutConfig
}

func (s *SwayLayout) processDevice(device *ipc.IpcInputDevice) bool {
	if s.Input != nil && device.Identifier != *s.Input || device.Type != "keyboard" {
		return false
	}
	s.currentLayoutIndex = device.XkbActiveLayoutIndex
	s.layouts = device.XkbLayoutNames
	return true
}

func (s *SwayLayout) Run(ch UpdateChan, ctx context.Context) {
	s.layoutLongToShort = rxkbcommon.GetLayoutShortNames()
	ipcClient, _ := ipc.NewIpcClient()
	defer ipcClient.Close()
	type IpcChanValue struct {
		typ     ipc.IpcMsgType
		payload interface{}
	}
	evc := make(chan IpcChanValue)
	s.ipc = ipcClient
	var err error
	var res interface{}
	ipcClient.SendRaw(ipc.IpcMsgTypeGetInputs, nil)
	_, res, err = ipcClient.Recv()
	if err != nil {
		Log.Print(err)
		return
	}
	for _, dev := range *res.(*[]ipc.IpcInputDevice) {
		if s.processDevice(&dev) {
			break
		}
	}
	ipcClient.SendRaw(ipc.IpcMsgTypeSubscribe, []byte(`["input"]`))
	if _, _, err = ipcClient.Recv(); err != nil {
		return
	}
	const throttleDuration = time.Millisecond * 50
	throttleTimer := time.NewTimer(throttleDuration)
	ch.SendUpdate()
	go func() {
		for {
			if t, d, err := ipcClient.Recv(); err == nil {
				evc <- IpcChanValue{t, d}
			} else {
				break
			}
		}
	}()
	for {
		select {
		case <-throttleTimer.C:
			ch.SendUpdate()
		case e := <-evc:
			throttleTimer.Reset(throttleDuration)
			if e.typ == ipc.IpcEventTypeInput {
				device := e.payload.(*ipc.InputChange).Input
				s.processDevice(&device)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *SwayLayout) OnEvent(e *I3barClickEvent, ctx context.Context) {
	if e.Button == ButtonRight {
		index := (s.currentLayoutIndex + 1) % len(s.layouts)
		cmd := fmt.Sprintf(`input type:keyboard xkb_switch_layout %d`, index)
		if err := s.ipc.SendRaw(ipc.IpcMsgTypeCommand, []byte(cmd)); err != nil {
			Log.Print(err)
		}
	}
}

func (s *SwayLayout) Render(cfg *AppConfig) []I3barBlock {
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
