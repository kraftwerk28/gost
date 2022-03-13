package blocks

import (
	"context"

	"github.com/godbus/dbus/v5"
	. "github.com/kraftwerk28/gost/core"
)

var objectManagerOutput map[dbus.ObjectPath](map[string](map[string]dbus.Variant))

// const nmDbusDest = "org.freedesktop.NetworkManager"
// const nmDbusBasePath dbus.ObjectPath = "/org/freedesktop/NetworkManager"
// const dbusGetProperty = "org.freedesktop.DBus.Properties.Get"

type BluezBlockConfig struct {
	DeviceFormat *ConfigFormat     `yaml:"device_format"`
	DeviceIcons  map[string]string `yaml:"device_icons"`
}

type BluezBlock struct {
	BluezBlockConfig
	dbus *dbus.Conn
}

func NewBluezBlock() I3barBlocklet {
	b := NetworkManagerBlock{}
	b.Format = NewConfigFormatFromString("{state_icon} {percentage*%}")
	b.propMap = map[string]interface{}{}
	return &b
}

func (t *BluezBlock) GetConfig() interface{} {
	return &t.BluezBlockConfig
}

func (t *BluezBlock) Run(ch UpdateChan, ctx context.Context) {
	b, err := dbus.SystemBus()
	if err != nil {
		Log.Print(err)
		return
	}
	defer b.Close()
	t.dbus = b
	c := make(chan *dbus.Signal)
	b.Signal(c)
	ch.SendUpdate()
	for {
		select {
		case <-ctx.Done():
			return
		case s := <-c:
			Log.Println(s)
			// s.Path -> path to AccessPoint
			// changedProps := s.Body[1].(map[string]dbus.Variant)
		}
	}
}

func (b *BluezBlock) Render() []I3barBlock {
	return nil
	// if b.state == nmStateConnectedGlobal {
	// 	c := b.connections[b.currentConnection]
	// 	ipMarshalled, _ := c.ipv4.MarshalText()
	// 	return []I3barBlock{{
	// 		FullText: b.Format.Expand(formatting.NamedArgs{
	// 			"ssid":        c.ssid,
	// 			"strength":    c.strength,
	// 			"ipv4":        string(ipMarshalled),
	// 			"status_icon": b.getStatusIcon(),
	// 		}),
	// 	}}
	// } else {
	// 	return []I3barBlock{{
	// 		FullText: b.Format.Expand(formatting.NamedArgs{
	// 			"status_icon": b.getStatusIcon(),
	// 		}),
	// 	}}
	// }
}

func init() {
	RegisterBlocklet("bluez", NewBluezBlock)
}
