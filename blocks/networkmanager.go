package blocks

import (
	"context"

	"github.com/godbus/dbus/v5"
	. "github.com/kraftwerk28/gost/core"
	"github.com/kraftwerk28/gost/core/formatting"
)

const dbusNetworkmanager = "org.freedesktop.NetworkManager"
const pathNetworkManager = "/org/freedesktop/NetworkManager"

type NetworkManagerBlockConfig struct {
	Format       *ConfigFormat     `yaml:"format"`
	UpowerDevice string            `yaml:"upower_device"`
	StateIcons   map[string]string `yaml:"state_icons"`
	LevelIcons   []string          `yaml:"level_icons"`
}

type NetworkManagerBlock struct {
	BatteryBlockConfig
	strength    int
	timeToEmpty int64
	state       uint32
	dbusConn    *dbus.Conn
	propMap     map[string]interface{}
}

func NewNetworkManagerBlock() I3barBlocklet {
	b := NetworkManagerBlock{}
	b.Format = NewConfigFormatFromString("{state_icon} {percentage}%")
	b.propMap = map[string]interface{}{
		"Strength": &b.strength,
	}
	return &b
}

func (t *NetworkManagerBlock) GetConfig() interface{} {
	return &t.BatteryBlockConfig
}

// func (t *BatteryBlock) listDevices(ctx context.Context) (r []dbus.ObjectPath, e error) {
// 	e = t.dbusConn.Object(
// 		dbUpower, pathUpower,
// 	).CallWithContext(
// 		ctx, "org.freedesktop.UPower.EnumerateDevices", 0,
// 	).Store(&r)
// 	if e != nil {
// 		Log.Println("Failed to get list")
// 	}
// 	return
// }

// func (t *BatteryBlock) loadInitial(b *dbus.Conn) {
// 	obj := b.Object(dbUpower, dbus.ObjectPath(t.UpowerDevice))
// 	iface := dbUpower + ".Device"
// 	for k, v := range t.propMap {
// 		obj.Call("org.freedesktop.DBus.Properties.Get", 0, iface, k).Store(v)
// 	}
// }

func (t *NetworkManagerBlock) Run(ch UpdateChan, ctx context.Context) {
	b, err := dbus.SystemBus()
	if err != nil {
		Log.Print(err)
		return
	}
	t.dbusConn = b
	if t.UpowerDevice == "" {
		p, err := t.findLaptopBattery(ctx)
		if err != nil {
			Log.Print(err)
			return
		}
		t.UpowerDevice = string(p)
	} else {
		t.UpowerDevice = "/org/freedesktop/UPower/devices/" + t.UpowerDevice
	}
	defer b.Close()
	if err := b.AddMatchSignalContext(
		ctx,
		dbus.WithMatchSender(dbUpower),
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		Log.Print(err)
		return
	}
	c := make(chan *dbus.Signal)
	b.Signal(c)
	t.loadInitial(b)
	ch.SendUpdate()
	for {
		select {
		case <-ctx.Done():
			return
		case s := <-c:
			changedProps := s.Body[1].(map[string]dbus.Variant)
			for k, v := range changedProps {
				if ref, ok := t.propMap[k]; ok {
					v.Store(ref)
				}
			}
			ch.SendUpdate()
		}
	}
}

func (t *NetworkManagerBlock) Render() []I3barBlock {
	return []I3barBlock{{
		FullText: t.Format.Expand(formatting.NamedArgs{
			"percentage":    t.strength,
			"time_to_empty": t.timeToEmpty,
		}),
	}}
}

func init() {
	RegisterBlocklet("networkmanager", NewNetworkManagerBlock)
}
