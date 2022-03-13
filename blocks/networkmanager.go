package blocks

import (
	"context"
	"encoding/binary"
	"net"

	"github.com/godbus/dbus/v5"
	. "github.com/kraftwerk28/gost/core"
	"github.com/kraftwerk28/gost/core/formatting"
)

const nmDbusDest = "org.freedesktop.NetworkManager"
const nmDbusBasePath dbus.ObjectPath = "/org/freedesktop/NetworkManager"
const dbusGetProperty = "org.freedesktop.DBus.Properties.Get"

type nmState uint32

const (
	nmStateAsleep          nmState = 10
	nmStateDisconnected            = 20
	nmStateDisconnecting           = 30
	nmStateConnecting              = 40
	nmStateConnectedLocal          = 50
	nmStateConnectedSite           = 60
	nmStateConnectedGlobal         = 70
)

type NetworkManagerBlockConfig struct {
	Format      *ConfigFormat     `yaml:"format"`
	PrimaryOnly bool              `yaml:"primary_only,omitempty"`
	StatusIcons map[string]string `yaml:"status_icons"`
}

type NmActiveConnection struct {
	connectionPath  dbus.ObjectPath
	accessPointPath dbus.ObjectPath
	ipv4ConfigPath  dbus.ObjectPath
	ipv6ConfigPath  dbus.ObjectPath
	strength        int
	ssid            string
	ipv4            net.IP
}

type NetworkManagerBlock struct {
	NetworkManagerBlockConfig
	dbus              *dbus.Conn
	propMap           map[string]interface{}
	connections       []NmActiveConnection
	state             nmState
	currentConnection int
}

func NewNetworkManagerBlock() I3barBlocklet {
	b := NetworkManagerBlock{}
	b.Format = NewConfigFormatFromString("{state_icon} {percentage*%}")
	b.propMap = map[string]interface{}{}
	return &b
}

func (t *NetworkManagerBlock) GetConfig() interface{} {
	return &t.NetworkManagerBlockConfig
}

func (t *NetworkManagerBlock) getDbusProperty(
	iface, prop string,
	out interface{},
) (err error) {
	obj := t.dbus.Object(nmDbusDest, nmDbusBasePath)
	err = obj.Call(dbusGetProperty, 0, string(iface), prop).Store(out)
	return
}

func (t *NetworkManagerBlock) loadState() (err error) {
	err = t.getDbusProperty(nmDbusDest, "State", &t.state)
	return
}

func (t *NetworkManagerBlock) loadConnectionProps(conn *NmActiveConnection) (err error) {
	apObj := t.dbus.Object(nmDbusDest, conn.accessPointPath)
	apIface := nmDbusDest + ".AccessPoint"
	if err = apObj.Call(dbusGetProperty, 0, apIface, "Strength").Store(&conn.strength); err != nil {
		return
	}
	var ssidByte []byte
	if err = apObj.Call(dbusGetProperty, 0, apIface, "Ssid").Store(&ssidByte); err != nil {
		return
	}
	conn.ssid = string(ssidByte)

	ip4Obj := t.dbus.Object(nmDbusDest, conn.ipv4ConfigPath)
	var ipv4Addresses [][]uint32
	if err = ip4Obj.Call(
		dbusGetProperty, 0,
		nmDbusDest+".IP4Config", "Addresses",
	).Store(&ipv4Addresses); err != nil {
		Log.Println(err)
		return
	}
	conn.ipv4 = make(net.IP, 4)
	binary.BigEndian.PutUint32(conn.ipv4, ipv4Addresses[0][0])
	return
}

func (t *NetworkManagerBlock) loadConnections() (err error) {
	if err = t.loadState(); err != nil {
		return
	}
	if t.state != nmStateConnectedGlobal {
		t.connections = nil
		return
	}
	t.connections = make([]NmActiveConnection, 0)
	var connPaths []dbus.ObjectPath
	if t.PrimaryOnly {
		var p dbus.ObjectPath
		if err = t.getDbusProperty(nmDbusDest, "PrimaryConnection", &p); err != nil {
			return
		}
		connPaths = []dbus.ObjectPath{p}
	} else {
		if err = t.getDbusProperty(nmDbusDest, "ActiveConnections", &connPaths); err != nil {
			return
		}
	}
	for _, connPath := range connPaths {
		var apPath, ipv4Path, ipv6Path dbus.ObjectPath
		dbusObj := t.dbus.Object(nmDbusDest, connPath)
		connIface := nmDbusDest + ".Connection.Active"
		if err = dbusObj.Call(dbusGetProperty, 0, connIface, "SpecificObject").Store(&apPath); err != nil {
			return
		}
		if err = dbusObj.Call(dbusGetProperty, 0, connIface, "Ip4Config").Store(&ipv4Path); err != nil {
			return
		}
		if err = dbusObj.Call(dbusGetProperty, 0, connIface, "Ip6Config").Store(&ipv6Path); err != nil {
			return
		}
		activeConn := NmActiveConnection{
			connectionPath:  connPath,
			accessPointPath: apPath,
			ipv4ConfigPath:  ipv4Path,
			ipv6ConfigPath:  ipv6Path,
		}
		if err = t.loadConnectionProps(&activeConn); err != nil {
			return
		}
		t.connections = append(t.connections, activeConn)
	}
	return
}

func (t *NetworkManagerBlock) addDbusSignals(ctx context.Context) (err error) {
	b := t.dbus
	if err = b.AddMatchSignalContext(
		ctx,
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchObjectPath("/org/freedesktop/NetworkManager"),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		return
	}
	if err = b.AddMatchSignalContext(
		ctx,
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchPathNamespace("/org/freedesktop/NetworkManager/ActiveConnection"),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		return
	}
	if err = b.AddMatchSignalContext(
		ctx,
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchPathNamespace("/org/freedesktop/NetworkManager/AccessPoint"),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		return
	}
	return
}

func (t *NetworkManagerBlock) getStatusIcon() string {
	switch t.state {
	case nmStateConnectedGlobal:
		return t.StatusIcons["connected"]
	case nmStateConnecting:
		return t.StatusIcons["connecting"]
	default:
		return t.StatusIcons["disconnected"]
	}
}

func (t *NetworkManagerBlock) Run(ch UpdateChan, ctx context.Context) {
	b, err := dbus.SystemBus()
	if err != nil {
		Log.Print(err)
		return
	}
	defer b.Close()
	t.dbus = b
	if err := t.addDbusSignals(ctx); err != nil {
		Log.Println(err)
		return
	}
	if err := t.loadConnections(); err != nil {
		// Don't fail
		Log.Print(err)
	}
	c := make(chan *dbus.Signal)
	b.Signal(c)
	ch.SendUpdate()
	for {
		select {
		case <-ctx.Done():
			return
		case s := <-c:
			// s.Path -> path to AccessPoint
			if len(t.connections) == 0 {
				// Was not connected, maybe now it is...
				t.loadConnections()
				ch.SendUpdate()
				continue
			}
			changedProps := s.Body[1].(map[string]dbus.Variant)
			shouldUpdate := false
			for i, conn := range t.connections {
				if s.Path == conn.accessPointPath {
					for k, v := range changedProps {
						switch k {
						case "Strength":
							v.Store(&t.connections[i].strength)
							shouldUpdate = true
						}
					}
				} else if s.Path == conn.connectionPath {
					if err := t.loadConnections(); err != nil {
						Log.Print(err)
					} else {
						ch.SendUpdate()
					}
				}
			}
			if shouldUpdate {
				ch.SendUpdate()
			}
		}
	}
}

func (b *NetworkManagerBlock) Render() []I3barBlock {
	if b.state == nmStateConnectedGlobal {
		c := b.connections[b.currentConnection]
		ipMarshalled, _ := c.ipv4.MarshalText()
		return []I3barBlock{{
			FullText: b.Format.Expand(formatting.NamedArgs{
				"ssid":        c.ssid,
				"strength":    c.strength,
				"ipv4":        string(ipMarshalled),
				"status_icon": b.getStatusIcon(),
			}),
		}}
	} else {
		return []I3barBlock{{
			FullText: b.Format.Expand(formatting.NamedArgs{
				"status_icon": b.getStatusIcon(),
			}),
		}}
	}
}

func init() {
	RegisterBlocklet("networkmanager", NewNetworkManagerBlock)
}
