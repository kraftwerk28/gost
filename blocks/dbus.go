package blocks

import (
	"context"

	"github.com/godbus/dbus/v5"
	. "github.com/kraftwerk28/gost/core"
)

type DbusConfig struct {
	ObjectPath  string `yaml:"object_path"`
	InitialText string `yaml:"initial_text"`
}

const dbusGetProperty = "org.freedesktop.DBus.Properties.Get"
const dbusPropertiesIface = "org.freedesktop.DBus.Properties"
const dbusCustomInterface = "com.kraftwerk28.gost"
const dbusObjectPath dbus.ObjectPath = "/org/freedesktop/DBus"

type DbusBlock struct {
	DbusConfig
	text string
}

func NewDbusBlock() I3barBlocklet {
	return &DbusBlock{}
}

type busObject struct {
	b  *DbusBlock
	ch UpdateChan
}

func (o *busObject) SetStatus(text string) *dbus.Error {
	o.b.text = text
	o.ch.SendUpdate()
	return nil
}

func (b *DbusBlock) Run(ch UpdateChan, ctx context.Context) {
	b.text = b.InitialText
	ch.SendUpdate()
	conn, err := dbus.SessionBus()
	if err != nil {
		Log.Println(err)
		return
	}
	defer conn.Close()
	dbusObject := &busObject{b, ch}
	if err := conn.Export(
		dbusObject,
		dbus.ObjectPath(b.ObjectPath),
		dbusCustomInterface,
	); err != nil {
		Log.Print(err)
		return
	}
	if _, err := conn.RequestName(
		dbusCustomInterface,
		dbus.NameFlagAllowReplacement,
	); err != nil {
		Log.Print(err)
		return
	}
	<-ctx.Done()
}

func (b *DbusBlock) GetConfig() interface{} {
	return &b.DbusConfig
}

func (b *DbusBlock) Render(cfg *AppConfig) []I3barBlock {
	return []I3barBlock{{
		FullText: b.text,
	}}
}

func init() {
	RegisterBlocklet("dbus", NewDbusBlock)
}
