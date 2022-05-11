package blocks

import (
	"context"
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
	. "github.com/kraftwerk28/gost/core"
)

type MprisBlockConfig struct {
}

type playbackStatus string

const (
	playbackStatusPlaying playbackStatus = "Playing"
	playbackStatusPaused  playbackStatus = "Paused"
	playbackStatusStopped playbackStatus = "Stopped"
)

type MprisPlayer struct {
	dbusName       string
	nameOwner      string
	metadata       map[string]dbus.Variant
	playbackStatus playbackStatus
}

type MprisBlock struct {
	MprisBlockConfig
	dbus    *dbus.Conn
	players []MprisPlayer
}

func NewMprisBlock() I3barBlocklet {
	return &MprisBlock{}
}

func (s *MprisBlock) GetConfig() interface{} {
	return &s.MprisBlockConfig
}

func (s *MprisBlock) AddPlayer(name string) {
}

const mprisPath dbus.ObjectPath = "/org/mpris/MediaPlayer2"
const mprisPlayerIface = "org.mpris.MediaPlayer2.Player"

func (b *MprisBlock) fetchPlayers() (err error) {
	var names []string
	if err = b.dbus.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names); err != nil {
		return
	}
	b.players = make([]MprisPlayer, 0)
	for _, name := range names {
		if strings.Contains(name, "org.mpris.MediaPlayer2") &&
			!strings.Contains(name, "playerctld") {
			p := b.dbus.Object(name, "/org/mpris/MediaPlayer2")
			player := MprisPlayer{}
			player.dbusName = name
			b.dbus.BusObject().Call(
				"org.freedesktop.DBus.GetNameOwner", 0, name,
			).Store(&player.nameOwner)
			p.Call(dbusGetProperty, 0, mprisPlayerIface, "PlaybackStatus").Store(&player.playbackStatus)
			p.Call(dbusGetProperty, 0, mprisPlayerIface, "Metadata").Store(&player.metadata)
			b.players = append(b.players, player)
		}
	}
	return
}

func (b *MprisBlock) Run(ch UpdateChan, ctx context.Context) {
	var err error
	b.dbus, err = dbus.ConnectSessionBus()
	if err != nil {
		Log.Print(err)
		return
	}
	defer b.dbus.Close()

	if err = b.fetchPlayers(); err != nil {
		Log.Print(err)
		return
	}
	ch.SendUpdate()

	err = b.dbus.AddMatchSignal(
		dbus.WithMatchObjectPath(mprisPath),
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
	)
	err = b.dbus.AddMatchSignal(
		dbus.WithMatchObjectPath(dbusObjectPath),
		dbus.WithMatchInterface("org.freedesktop.DBus"),
	)

	if err != nil {
		Log.Print(err)
		return
	}

	c := make(chan *dbus.Signal)
	b.dbus.Signal(c)

	for {
		select {
		case sig := <-c:
			switch sig.Path {
			case dbusObjectPath:
				b.fetchPlayers()
				ch.SendUpdate()
			case mprisPath:
				var player *MprisPlayer
				shouldUpdate := false
				for i := range b.players {
					player = &b.players[i]
					if player.dbusName == sig.Sender || player.nameOwner == sig.Sender {
						break
					}
				}
				if player == nil {
					break
				}
				chProps := sig.Body[1].(map[string]dbus.Variant)
				if p, ok := chProps["PlaybackStatus"]; ok {
					p.Store(&player.playbackStatus)
					shouldUpdate = true
				}
				if p, ok := chProps["Metadata"]; ok {
					p.Store(&player.metadata)
					shouldUpdate = true
				}
				if shouldUpdate {
					ch.SendUpdate()
				}
			}
		case <-ctx.Done():
			return
		}
	}

}

func (t *MprisBlock) OnEvent(e *I3barClickEvent, ctx context.Context) {
}

func (t *MprisBlock) Render(cfg *AppConfig) []I3barBlock {
	s := make([]string, 0)
	for _, pl := range t.players {
		var icon string
		switch pl.playbackStatus {
		case playbackStatusPaused:
			icon = " "
		case playbackStatusPlaying:
			icon = " "
		case playbackStatusStopped:
			icon = " "
		}
		s = append(s, fmt.Sprintf("%s[%s]", icon, pl.metadata["xesam:title"]))
	}
	return []I3barBlock{{
		FullText: strings.Join(s, " | "),
	}}
}

func init() {
	RegisterBlocklet("mpris", NewMprisBlock)
}
