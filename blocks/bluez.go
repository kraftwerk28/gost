package blocks

import (
	"context"
	"strings"

	"github.com/godbus/dbus/v5"
	. "github.com/kraftwerk28/gost/core"
	"github.com/kraftwerk28/gost/core/formatting"
)

type bluezObjectManagerOutput map[dbus.ObjectPath](map[string](map[string]dbus.Variant))

// Displays connected bluetooth devices
type BluezBlockConfig struct {
	BaseBlockletConfig `yaml:",inline"`
	// Mac address of the device
	Device             string            `yaml:"mac"`
	Format             *ConfigFormat     `yaml:"format"`
	// Device format
	DeviceFormat       *ConfigFormat     `yaml:"device_format"`
	Icons              map[string]string `yaml:"icons"`
	ExcludeMac         []string          `yaml:"exclude"`
}

type bluezDevice struct {
	path              dbus.ObjectPath
	connected         bool
	name, alias, icon string
}

type BluezBlock struct {
	BluezBlockConfig
	dbus    *dbus.Conn
	devices map[dbus.ObjectPath]*bluezDevice
}

func NewBluezBlock() I3barBlocklet {
	b := BluezBlock{}
	b.DeviceFormat = NewConfigFormatFromString("{icon}")
	return &b
}

func (t *BluezBlock) GetConfig() interface{} {
	return &t.BluezBlockConfig
}

func (b *BluezBlock) isExcluded(addr string) bool {
	for _, p := range b.ExcludeMac {
		if p == addr {
			return true
		}
	}
	return false
}

func (b *BluezBlock) loadDevices() (err error) {
	var bluezObjects bluezObjectManagerOutput
	if err = b.dbus.Object("org.bluez", "/").Call(
		"org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0,
	).Store(&bluezObjects); err != nil {
		return
	}
	b.devices = make(map[dbus.ObjectPath]*bluezDevice)
	for path, v := range bluezObjects {
		if info, ok := v["org.bluez.Device1"]; ok {
			var addr, name, alias, icon string
			var connected bool
			info["Address"].Store(&addr)
			if b.isExcluded(addr) {
				continue
			}
			if i, ok := info["Icon"]; ok {
				i.Store(&icon)
			}
			info["Name"].Store(&name)
			info["Alias"].Store(&alias)
			info["Connected"].Store(&connected)
			b.devices[path] = &bluezDevice{path, connected, name, alias, icon}
		}
	}
	return
}

func (t *BluezBlock) addSignals() (err error) {
	b := t.dbus
	if err = b.AddMatchSignal(
		dbus.WithMatchPathNamespace("/org/bluez"),
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
		dbus.WithMatchArg(0, "org.bluez.Device1"),
	); err != nil {
		return
	}
	if err = b.AddMatchSignal(
		dbus.WithMatchObjectPath("/"),
		dbus.WithMatchInterface("org.freedesktop.DBus.ObjectManager"),
		dbus.WithMatchMember("InterfaceAdded"),
	); err != nil {
		return
	}
	if err = b.AddMatchSignal(
		dbus.WithMatchObjectPath("/"),
		dbus.WithMatchInterface("org.freedesktop.DBus.ObjectManager"),
		dbus.WithMatchMember("InterfaceRemoved"),
	); err != nil {
		return
	}
	return
}

func (t *BluezBlock) Run(ch UpdateChan, ctx context.Context) {
	b, err := dbus.ConnectSystemBus()
	if err != nil {
		Log.Print(err)
		return
	}
	defer b.Close()
	t.dbus = b
	if err := t.addSignals(); err != nil {
		Log.Print(err)
		return
	}
	if err := t.loadDevices(); err != nil {
		Log.Print(err)
	} else {
		ch.SendUpdate()
	}
	c := make(chan *dbus.Signal)
	b.Signal(c)
	for {
		select {
		case <-ctx.Done():
			return
		case s := <-c:
			switch s.Name {
			case "org.freedesktop.DBus.ObjectManager.InterfaceAdded",
				"org.freedesktop.DBus.ObjectManager.InterfaceRemoved":
				t.loadDevices()
				ch.SendUpdate()
			case "org.freedesktop.DBus.Properties.PropertiesChanged":
				if t.devices[s.Path] == nil {
					t.loadDevices()
					ch.SendUpdate()
					break
				}
				props := s.Body[1].(map[string]dbus.Variant)
				if conn, ok := props["Connected"]; ok {
					var connected bool
					conn.Store(&connected)
					t.devices[s.Path].connected = connected
					ch.SendUpdate()
				}
			}
		}
	}
}

func (b *BluezBlock) Render(cfg *AppConfig) []I3barBlock {
	labels := []string{}
	for _, d := range b.devices {
		if d.connected {
			args := formatting.NamedArgs{
				"icon":  b.Icons[d.icon],
				"name":  d.name,
				"alias": d.alias,
			}
			labels = append(labels, b.DeviceFormat.Expand(args))
		}
	}
	return []I3barBlock{{
		FullText: b.Format.Expand(formatting.NamedArgs{
			"devices": strings.Join(labels, " "),
		}),
	}}
}

func init() {
	RegisterBlocklet("bluez", NewBluezBlock)
}
