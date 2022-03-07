package blocks

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	. "github.com/kraftwerk28/gost/core"
)

type BatteryBlockConfig struct {
	Device string `yaml:"device"`
}

type BatteryBlock struct {
	BatteryBlockConfig
	percentage float32
}

func NewBatteryBlock() I3barBlocklet {
	return &BatteryBlock{}
}

func (t *BatteryBlock) GetConfig() interface{} {
	return &t.BatteryBlockConfig
}

const dbUpower = "org.freedesktop.UPower"

func (t *BatteryBlock) refresh(b *dbus.Conn) {
	if err := b.Object(
		dbUpower,
		dbus.ObjectPath("/org/freedesktop/UPower/devices/"+t.Device),
	).Call(
		"org.freedesktop.DBus.Properties.Get", 0,
		dbUpower+".Device", "Percentage",
	).Store(&t.percentage); err != nil {
		Log.Print(err)
	}
}

func (t *BatteryBlock) Run(ch UpdateChan, ctx context.Context) {
	b, _ := dbus.SystemBus()
	if err := b.AddMatchSignal(
		dbus.WithMatchSender(dbUpower),
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		Log.Print(err)
	}
	c := make(chan *dbus.Signal)
	b.Signal(c)
	t.refresh(b)
	ch.SendUpdate()
	for s := range c {
		Log.Printf("%+v\n", s)
		t.refresh(b)
		ch.SendUpdate()
	}
}

func (t *BatteryBlock) Render() []I3barBlock {
	return []I3barBlock{{
		FullText: fmt.Sprintf("%.1f", t.percentage),
	}}
}

func init() {
	RegisterBlocklet("battery", NewBatteryBlock)
}
