package blocks

import (
	"context"

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
	Format      *ConfigFormat `yaml:"format"`
	PrimaryOnly bool          `yaml:"primary_only,omitempty"`
}

type NmActiveConnection struct {
	connectionPath  dbus.ObjectPath
	accessPointPath dbus.ObjectPath
	strength        int
	ssid            string
}

type NetworkManagerBlock struct {
	NetworkManagerBlockConfig
	dbus        *dbus.Conn
	propMap     map[string]interface{}
	connections []NmActiveConnection
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

func (t *NetworkManagerBlock) getState() (result nmState, err error) {
	err = t.getDbusProperty(nmDbusDest, "State", &result)
	return
}

func (t *NetworkManagerBlock) loadConnectionProps(conn *NmActiveConnection) (err error) {
	apObj := t.dbus.Object(nmDbusDest, conn.accessPointPath)
	apIface := nmDbusDest + ".AccessPoint"
	if err = apObj.Call(
		dbusGetProperty, 0,
		apIface, "Strength",
	).Store(&conn.strength); err != nil {
		return
	}
	var ssidByte []byte
	if err = apObj.Call(
		dbusGetProperty, 0,
		apIface, "Ssid",
	).Store(&ssidByte); err != nil {
		return
	}
	conn.ssid = string(ssidByte)
	return
}

func (t *NetworkManagerBlock) loadConnections(primaryOnly bool) (err error) {
	t.connections = make([]NmActiveConnection, 0)
	var connPaths []dbus.ObjectPath
	if primaryOnly {
		var p dbus.ObjectPath
		if err = t.getDbusProperty(
			nmDbusDest, "PrimaryConnection", &p,
		); err != nil {
			return
		}
		connPaths = []dbus.ObjectPath{p}
	} else {
		if err = t.getDbusProperty(
			nmDbusDest, "ActiveConnections", &connPaths,
		); err != nil {
			return
		}
	}
	for _, connPath := range connPaths {
		var apPath dbus.ObjectPath
		if err = t.dbus.Object(
			nmDbusDest, connPath,
		).Call(
			dbusGetProperty, 0,
			nmDbusDest+".Connection.Active", "SpecificObject",
		).Store(&apPath); err != nil {
			return
		}
		activeConn := NmActiveConnection{
			connectionPath:  connPath,
			accessPointPath: apPath,
		}
		t.loadConnectionProps(&activeConn)
		t.connections = append(t.connections, activeConn)
	}
	return
}

func (t *NetworkManagerBlock) Run(ch UpdateChan, ctx context.Context) {
	b, err := dbus.SystemBus()
	if err != nil {
		Log.Print(err)
		return
	}
	defer b.Close()
	t.dbus = b
	if err := t.loadConnections(t.PrimaryOnly); err != nil {
		return
	}
	if err := b.AddMatchSignalContext(
		ctx,
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchObjectPath("/org/freedesktop/NetworkManager"),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		Log.Print(err)
		return
	}
	if err := b.AddMatchSignalContext(
		ctx,
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchPathNamespace("/org/freedesktop/NetworkManager/ActiveConnection"),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		Log.Print(err)
		return
	}
	if err := b.AddMatchSignalContext(
		ctx,
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchPathNamespace("/org/freedesktop/NetworkManager/AccessPoint"),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		Log.Print(err)
		return
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
			// Log.Printf("NetworkManager signal: %+v\n", s)
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
					// TODO
				}
			}
			if shouldUpdate {
				ch.SendUpdate()
			}
			// // for k, v := range changedProps {
			// // 	if ref, ok := t.propMap[k]; ok {
			// // 		v.Store(ref)
			// // 	}
			// // }
			// ch.SendUpdate()
		}
	}
}

func (t *NetworkManagerBlock) Render() []I3barBlock {
	if len(t.connections) == 0 {
		return nil
	}
	conn := t.connections[0]
	return []I3barBlock{{
		FullText: t.Format.Expand(formatting.NamedArgs{
			"ssid":     conn.ssid,
			"strength": conn.strength,
		}),
	}}
}

func init() {
	RegisterBlocklet("networkmanager", NewNetworkManagerBlock)
}
