package blocks

import (
	"context"

	"github.com/kraftwerk28/gost/blocks/ipc"
	. "github.com/kraftwerk28/gost/core"
)

type SwayWindowConfig struct {
}

type SwayWindow struct {
	name string
}

func NewSwayWindow() I3barBlocklet {
	return &SwayWindow{}
}

// func (s *SwayWindow) GetConfig() interface{} {
// 	return nil
// }

func (t *SwayWindow) Run(ch UpdateChan, ctx context.Context) {
	client, _ := ipc.NewIpcClient()
	client.SendRaw(ipc.IpcMsgTypeSubscribe, []byte(`["window"]`))
	client.Recv()
	for {
		_, m, _ := client.Recv()
		t.name = m.(*ipc.WindowChange).Container.Name
		ch.SendUpdate()
	}
}

func (t *SwayWindow) Render(cfg *AppConfig) []I3barBlock {
	return []I3barBlock{{FullText: t.name}}
}

func init() {
	RegisterBlocklet("sway_window", NewSwayWindow)
}
