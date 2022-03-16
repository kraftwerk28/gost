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
	defer ipcClient.Close()
	type IpcChanValue struct {
		typ     ipc.IpcMsgType
		payload interface{}
	}
	evc := make(chan IpcChanValue)
	go func() {
		for {
			t, d, err := ipcClient.Recv()
			if err != nil {
				break
			}
			evc <- IpcChanValue{t, d}
		}
	}()
	s.ipc = ipcClient
	var err error
	var res interface{}
	ipcClient.SendRaw(ipc.IpcMsgTypeGetInputs, nil)
	_, res, err = ipcClient.Recv()
	if err != nil {
		Log.Print(err)
		return
	}
	for _, device := range *res.(*[]ipc.IpcInputDevice) {
		if device.Type == "keyboard" {
			s.currentLayoutIndex = device.XkbActiveLayoutIndex
			s.layouts = device.XkbLayoutNames
			ch.SendUpdate()
			break
		}
	}
	ipcClient.SendRaw(ipc.IpcMsgTypeSubscribe, []byte(`["input"]`))
	if _, _, err = ipcClient.Recv(); err != nil {
		return
	}
	const throttleDuration = time.Millisecond * 50
	throttleTimer := time.NewTimer(throttleDuration)
	for {
		select {
		case <-throttleTimer.C:
			ch.SendUpdate()
		case e := <-evc:
			throttleTimer.Reset(throttleDuration)
			if e.typ == ipc.IpcEventTypeInput {
				device := e.payload.(*ipc.InputChange).Input
				if device.Type == "keyboard" {
					s.currentLayoutIndex = device.XkbActiveLayoutIndex
					s.layouts = device.XkbLayoutNames
				}
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
