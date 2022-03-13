package blocks

import (
	"context"

	"github.com/godbus/dbus/v5"
	. "github.com/kraftwerk28/gost/core"
	"github.com/kraftwerk28/gost/core/formatting"
)

const upowerDbusDest = "org.freedesktop.UPower"
const upowerDbusBasePath dbus.ObjectPath = "/org/freedesktop/UPower"

const (
	upowerStateUnknown          uint32 = 0
	upowerStateCharging                = 1
	upowerStateDischarging             = 2
	upowerStateEmpty                   = 3
	upowerStateFullyCharged            = 4
	upowerStatePendingCharge           = 5
	upowerStatePendingDischarge        = 6
)

const (
	deviceTypeBattery  uint32 = 2
	deviceTypeKeyboard        = 6
)

type BatteryBlockConfig struct {
	Format       *ConfigFormat     `yaml:"format"`
	UpowerDevice string            `yaml:"upower_device"`
	StateIcons   map[string]string `yaml:"state_icons"`
	LevelIcons   []string          `yaml:"level_icons"`
}

type BatteryBlock struct {
	BatteryBlockConfig
	percentage  int
	timeToEmpty int64
	state       uint32
	dbusConn    *dbus.Conn
	propMap     map[string]interface{}
}

func NewBatteryBlock() I3barBlocklet {
	b := BatteryBlock{}
	b.Format = NewConfigFormatFromString("{state_icon} {percentage}%")
	b.propMap = map[string]interface{}{
		"Percentage":  &b.percentage,
		"TimeToEmpty": &b.timeToEmpty,
		"State":       &b.state,
	}
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

func (t *BatteryBlock) listDevices(ctx context.Context) (r []dbus.ObjectPath, e error) {
	e = t.dbusConn.Object(
		upowerDbusDest, upowerDbusBasePath,
	).CallWithContext(
		ctx, "org.freedesktop.UPower.EnumerateDevices", 0,
	).Store(&r)
	if e != nil {
		Log.Println("Failed to get list")
	}
	return
}

func (t *BatteryBlock) findLaptopBattery(ctx context.Context) (p dbus.ObjectPath, e error) {
	var devices []dbus.ObjectPath
	if devices, e = t.listDevices(ctx); e != nil {
		return
	}
	for _, oPath := range devices {
		var devType uint32
		e = t.dbusConn.Object(
			upowerDbusDest, oPath,
		).CallWithContext(
			ctx, "org.freedesktop.DBus.Properties.Get", 0,
			upowerDbusDest+".Device", "Type",
		).Store(&devType)
		if e != nil {
			return
		}
		if devType == deviceTypeBattery {
			p = oPath
			return
		}
	}
	return
}

func (t *BatteryBlock) loadInitial(b *dbus.Conn) {
	obj := b.Object(upowerDbusDest, dbus.ObjectPath(t.UpowerDevice))
	iface := upowerDbusDest + ".Device"
	for k, v := range t.propMap {
		obj.Call("org.freedesktop.DBus.Properties.Get", 0, iface, k).Store(v)
	}
}

func (t *BatteryBlock) Run(ch UpdateChan, ctx context.Context) {
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
		dbus.WithMatchSender(upowerDbusDest),
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

func (t *BatteryBlock) Render() []I3barBlock {
	args := formatting.NamedArgs{
		"percentage":    t.percentage,
		"time_to_empty": t.timeToEmpty,
		"state_icon":    t.getStateIcon(),
	}
	if t.state == upowerStateCharging {
		args["is_charging"] = t.StateIcons["charging"]
	}
	return []I3barBlock{{FullText: t.Format.Expand(args)}}
}

func init() {
	RegisterBlocklet("battery", NewBatteryBlock)
}
