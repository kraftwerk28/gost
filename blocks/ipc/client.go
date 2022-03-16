package ipc

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"os"
)

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
			out = &IpcResult{}
		case IpcMsgTypeCommand:
			out = &IpcCmdResult{}
		}
	}
	if err := json.Unmarshal(resBody, out); err != nil {
		return IpcMsgTypeInvalid, nil, err
	}
	return resHeader.Typ, out, nil
}

func (s *IpcClient) Close() error {
	return s.conn.Close()
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
