package ipc

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"os"
)

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

type IpcClient struct {
	conn net.Conn
}

func NewIpcClient() (*IpcClient, error) {
	l, err := net.Dial("unix", os.Getenv("SWAYSOCK"))
	if err != nil {
		return nil, err
	}
	return &IpcClient{l}, nil
}

func (s *IpcClient) Recv() (IpcMsgType, interface{}, error) {
	resHeader := &IpcHeader{}
	if err := binary.Read(s.conn, binary.LittleEndian, resHeader); err != nil {
		return IpcMsgTypeInvalid, nil, err
	}
	resBody := make([]byte, resHeader.Len)
	if err := binary.Read(s.conn, binary.LittleEndian, resBody); err != nil {
		return IpcMsgTypeInvalid, nil, err
	}
	var out interface{}
	if resHeader.Typ&0x80000000 != 0 {
		resHeader.Typ &= 0x7fffffff
		switch resHeader.Typ {
		case IpcEventTypeInput:
			out = &InputChange{}
		case IpcEventTypeWindow:
			out = &WindowChange{}
		}
	} else {
		switch resHeader.Typ {
		case IpcMsgTypeGetInputs:
			out = &[]IpcInputDevice{}
		case IpcMsgTypeSubscribe:
			out = &IpcSubscribeResult{}
		}
	}
	if err := json.Unmarshal(resBody, out); err != nil {
		return IpcMsgTypeInvalid, nil, err
	}
	return resHeader.Typ, out, nil
}

func (s *IpcClient) SendRaw(t IpcMsgType, data []byte) error {
	var l uint32
	if data != nil {
		l = uint32(len(data))
	}
	if err := binary.Write(
		s.conn,
		binary.LittleEndian,
		&IpcHeader{ipcMagic, l, t},
	); err != nil {
		return err
	}
	if l > 0 {
		if err := binary.Write(s.conn, binary.LittleEndian, data); err != nil {
			return err
		}
	}
	return nil
}

func (s *IpcClient) Send(t IpcMsgType, data interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.SendRaw(t, raw)
}
