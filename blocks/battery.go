package blocks

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	. "github.com/kraftwerk28/gost/core"
	"github.com/kraftwerk28/gost/core/formatting"
)

const upowerDbusDest = "org.freedesktop.UPower"
const upowerDbusBasePath dbus.ObjectPath = "/org/freedesktop/UPower"

const (
	upowerStateUnknown          uint32 = 0
	upowerStateCharging         uint32 = 1
	upowerStateDischarging      uint32 = 2
	upowerStateEmpty            uint32 = 3
	upowerStateFullyCharged     uint32 = 4
	upowerStatePendingCharge    uint32 = 5
	upowerStatePendingDischarge uint32 = 6
)

const (
	deviceTypeBattery  uint32 = 2
	deviceTypeKeyboard uint32 = 6
)

// Display battery charge level. Requires UPower to work over DBus
// `upower_device` can be obtained via DBus using the following command:
// ```bash
//
//	$ busctl --json=pretty --system call org.freedesktop.UPower \
//	  /org/freedesktop/UPower org.freedesktop.UPower \
//	  EnumerateDevices \
//	  | jq -r '.data[0] | map(match("[^/]+$").string)[]'`
//
// ```
// If empty, the program will try to detect battery device.
type BatteryBlockConfig struct {
	Format *ConfigFormat `yaml:"format"`
	// Device name. See above.
	UpowerDevice string            `yaml:"upower_device"`
	StateIcons   map[string]string `yaml:"state_icons"`
	LevelIcons   []string          `yaml:"level_icons"`
	UrgentLevel  *int              `yaml:"urgent_level"`
}

type BatteryBlock struct {
	BatteryBlockConfig
	percentage  int
	timeToEmpty int64
	state       uint32
	dbusConn    *dbus.Conn
	propMap     map[string]interface{}
	available   bool
}

func NewBatteryBlock() I3barBlocklet {
	b := BatteryBlock{}
	b.Format = NewConfigFormatFromString("{state_icon} {percentage}%")
	b.propMap = map[string]interface{}{
		"Percentage":  &b.percentage,
		"TimeToEmpty": &b.timeToEmpty,
		"State":       &b.state,
	}
	b.available = false
	return &b
}

func (t *BatteryBlock) GetConfig() interface{} {
	return &t.BatteryBlockConfig
}

func (t *BatteryBlock) getLevelIcon() string {
	x := float64(t.percentage) / 100.1 * float64(len(t.LevelIcons))
	return t.LevelIcons[int(x)]
}

func (t *BatteryBlock) getStateIcon() string {
	switch t.state {
	case upowerStateUnknown:
		return t.StateIcons["unknown"]
	case upowerStateCharging:
		if icon, ok := t.StateIcons["discharging"]; ok {
			return icon
		}
		return t.getLevelIcon()
	case upowerStateDischarging:
		if icon, ok := t.StateIcons["discharging"]; ok {
			return icon
		}
		return t.getLevelIcon()
	case upowerStateEmpty:
		return t.StateIcons["empty"]
	case upowerStateFullyCharged:
		if icon, ok := t.StateIcons["fully_charged"]; ok {
			return icon
		}
		return t.getLevelIcon()
	case upowerStatePendingCharge:
		// TODO: how to handle this states?
		return t.StateIcons["pending_charge"]
	case upowerStatePendingDischarge:
		// TODO: how to handle this states?
		return t.StateIcons["pending_discharge"]
	default:
		return ""
	}
}

func (t *BatteryBlock) listDevices(ctx context.Context) ([]dbus.ObjectPath, error) {
	var paths []dbus.ObjectPath
	o := t.dbusConn.Object(upowerDbusDest, upowerDbusBasePath)
	if err := o.CallWithContext(ctx, "org.freedesktop.UPower.EnumerateDevices", 0).Store(&paths); err != nil {
		Log.Println("Failed to get list")
		return nil, err
	}
	return paths, nil
}

func (t *BatteryBlock) findLaptopBattery(ctx context.Context) (p dbus.ObjectPath, e error) {
	var devices []dbus.ObjectPath
	if devices, e = t.listDevices(ctx); e != nil {
		return
	}
	for _, oPath := range devices {
		var devType uint32
		o := t.dbusConn.Object(upowerDbusDest, oPath)
		if err := o.CallWithContext(ctx, dbusGetProperty, 0, upowerDbusDest+".Device", "Type").Store(&devType); err != nil {
			return
		}
		if devType == deviceTypeBattery {
			p = oPath
			return
		}
	}
	return
}

func (t *BatteryBlock) loadInitial() (err error) {
	obj := t.dbusConn.Object(upowerDbusDest, dbus.ObjectPath(t.UpowerDevice))
	iface := upowerDbusDest + ".Device"
	for k, v := range t.propMap {
		if err = obj.Call(dbusGetProperty, 0, iface, k).Store(v); err != nil {
			return
		}
	}
	return
}

func (t *BatteryBlock) Run(ch UpdateChan, ctx context.Context) {
	b, err := dbus.ConnectSystemBus()
	if err != nil {
		Log.Println("ConnectSystemBus", err)
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
		dbus.WithMatchObjectPath(dbus.ObjectPath(t.UpowerDevice)),
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		Log.Print(err)
		return
	}
	if err := b.AddMatchSignalContext(
		ctx,
		dbus.WithMatchObjectPath(upowerDbusBasePath),
		dbus.WithMatchInterface(upowerDbusDest),
		dbus.WithMatchMember("DeviceAdded"),
	); err != nil {
		Log.Print(err)
		return
	}
	if err := b.AddMatchSignalContext(
		ctx,
		dbus.WithMatchObjectPath(upowerDbusBasePath),
		dbus.WithMatchInterface(upowerDbusDest),
		dbus.WithMatchMember("DeviceRemoved"),
	); err != nil {
		Log.Print(err)
		return
	}

	c := make(chan *dbus.Signal)
	b.Signal(c)
	if err := t.loadInitial(); err != nil {
		t.available = false
		Log.Print(err)
	} else {
		t.available = true
		ch.SendUpdate()
	}
	for {
		select {
		case <-ctx.Done():
			return
		case s := <-c:
			if s.Path == dbus.ObjectPath(t.UpowerDevice) {
				if !t.available {
					continue
				}
				changedProps := s.Body[1].(map[string]dbus.Variant)
				for k, v := range changedProps {
					if ref, ok := t.propMap[k]; ok {
						v.Store(ref)
					}
				}
				ch.SendUpdate()
			} else if s.Path == upowerDbusBasePath &&
				s.Name == "org.freedesktop.UPower.DeviceAdded" &&
				dbus.ObjectPath(t.UpowerDevice) == s.Body[0].(dbus.ObjectPath) {
				if err := t.loadInitial(); err != nil {
					Log.Print(err)
				} else {
					t.available = true
					ch.SendUpdate()
				}
			} else if s.Path == upowerDbusBasePath &&
				s.Name == "org.freedesktop.UPower.DeviceRemoved" &&
				dbus.ObjectPath(t.UpowerDevice) == s.Body[0].(dbus.ObjectPath) {
				t.available = false
				ch.SendUpdate()
			}
		}
	}
}

func (t *BatteryBlock) Render(cfg *AppConfig) []I3barBlock {
	if !t.available {
		return nil
	}
	args := formatting.NamedArgs{
		"percentage":    t.percentage,
		"time_to_empty": t.timeToEmpty,
		"state_icon": fmt.Sprintf(
			`<span color="%v">%s</span>`,
			cfg.Theme.HSVColor(PercentageToHue(t.percentage)),
			t.getStateIcon(),
		),
	}
	if t.state == upowerStateCharging {
		args["is_charging"] = t.StateIcons["charging"]
	}
	b := I3barBlock{FullText: t.Format.Expand(args)}
	if t.UrgentLevel != nil && t.percentage <= *t.UrgentLevel {
		b.Urgent = true
	}
	b.Markup = MarkupPango
	return []I3barBlock{b}
}

func init() {
	RegisterBlocklet("battery", NewBatteryBlock)
}
