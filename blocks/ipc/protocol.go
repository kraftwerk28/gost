package ipc

type IpcMsgType uint32

const (
	IpcMsgTypeCommand   IpcMsgType = 0
	IpcMsgTypeSubscribe IpcMsgType = 2
	IpcMsgTypeGetInputs IpcMsgType = 100
	IpcEventTypeInput   IpcMsgType = 0x15
	IpcEventTypeWindow  IpcMsgType = 0x3
	IpcMsgTypeInvalid   IpcMsgType = 0x7fffffff
)

type IpcHeader struct {
	Magic [6]byte
	Len   uint32
	Typ   IpcMsgType
}

var ipcMagic = [6]byte{'i', '3', '-', 'i', 'p', 'c'}

type IpcInputDevice struct {
	Type                 string   `json:"type"`
	XkbActiveLayoutName  string   `json:"xkb_active_layout_name"`
	XkbLayoutNames       []string `json:"xkb_layout_names"`
	XkbActiveLayoutIndex int      `json:"xkb_active_layout_index"`
}

type IpcContainer struct {
	Name  string `json:"name"`
	AppId string `json:"app_id"`
}

type InputChange struct {
	Change string         `json:"change"`
	Input  IpcInputDevice `json:"input"`
}

type WindowChange struct {
	Change    string       `json:"change"`
	Container IpcContainer `json:"container"`
}

type IpcSubscribeResult struct {
	Success bool `json:"success"`
}
