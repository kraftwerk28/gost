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

// Displays current keyboard layout.
// Uses sway's IPC API for retrieving the info (i.e. won't work with i3wm)
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

	ipcClient, err := ipc.NewIpcClient()
	// s.ipc, err := ipc.NewIpcClient()
	if err != nil {
		return
	}
	defer ipcClient.Close()
	s.ipc = ipcClient

	var swayInputs interface{}
	var msgType ipc.IpcMsgType
	ipcClient.SendRaw(ipc.IpcMsgTypeGetInputs, nil)
	msgType, swayInputs, err = ipcClient.Recv()
	if err != nil {
		Log.Print(err)
		return
	}
	if msgType != ipc.IpcMsgTypeGetInputs {
		Log.Println("Invalid IPC response")
		return
	}
	for _, dev := range *swayInputs.(*[]ipc.IpcInputDevice) {
		if s.processDevice(&dev) {
			break
		}
	}

	ipcClient.Send(ipc.IpcMsgTypeSubscribe, []string{"input"})
	if msgType, _, err = ipcClient.Recv(); err != nil {
		return
	}

	ch.SendUpdate()

	type IpcChanValue struct {
		typ     ipc.IpcMsgType
		payload interface{}
	}
	evc := make(chan IpcChanValue)
	go func() {
		for {
			if t, d, err := ipcClient.Recv(); err == nil {
				evc <- IpcChanValue{t, d}
			} else {
				break
			}
		}
	}()
	throttleTimer := time.NewTimer(time.Millisecond * 50)
	for {
		select {
		case <-throttleTimer.C:
			ch.SendUpdate()
		case e := <-evc:
			throttleTimer.Reset(throttleDuration)
			switch e.typ {
			case ipc.IpcEventTypeInput:
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
