package blocks

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/godbus/dbus/v5"
	. "github.com/kraftwerk28/gost/core"
	"github.com/kraftwerk28/gost/core/formatting"
)

const nmDbusDest = "org.freedesktop.NetworkManager"
const nmDbusBasePath dbus.ObjectPath = "/org/freedesktop/NetworkManager"

type NetworkManagerBlockConfig struct {
	BaseBlockletConfig `yaml:",inline"`
	Format             *ConfigFormat          `yaml:"format"`
	AccessPointFormat  *ConfigFormat          `yaml:"ap_format"`
	PrimaryOnly        bool                   `yaml:"primary_only"`
	Icons              map[string]interface{} `yaml:"icons"`
}

type NmActiveConnection struct {
	objectPath dbus.ObjectPath
	device     NmDevice
	ip4Config  NmIp4Config
	isVpn      bool
}

func (c *NmActiveConnection) isWireless() bool {
	switch c.device.deviceType {
	case NM_DEVICE_TYPE_WIFI, NM_DEVICE_TYPE_WIFI_P2P:
		return true
	default:
		return false
	}
}

func (c *NmActiveConnection) loadProps(conn *dbus.Conn) (err error) {
	o := conn.Object(nmDbusDest, c.objectPath)
	iface := nmDbusDest + ".Connection.Active"
	var devices []dbus.ObjectPath
	if err = dbusGetProp(o, iface, "Devices", &devices); err != nil {
		return
	}
	if len(devices) == 0 {
		return
	}
	c.device.objectPath = devices[0]
	if err = c.device.loadProps(conn); err != nil {
		return
	}
	if err = dbusGetProp(o, iface, "Ip4Config", &c.ip4Config.objectPath); err != nil {
		return
	}
	if err = c.ip4Config.loadProps(conn); err != nil {
		return
	}
	if err = dbusGetProp(o, iface, "Vpn", &c.isVpn); err != nil {
		return
	}
	return
}

type NmDevice struct {
	objectPath  dbus.ObjectPath
	deviceType  nmDeviceType
	accessPoint NmAccessPoint
	bitrate     uint32
}

func (d *NmDevice) loadProps(conn *dbus.Conn) (err error) {
	o := conn.Object(nmDbusDest, d.objectPath)
	if err = dbusGetProp(o, nmDbusDest+".Device", "DeviceType", &d.deviceType); err != nil {
		return
	}
	dbusGetProp(o, nmDbusDest+".Device.Wireless", "Bitrate", &d.bitrate)
	dbusGetProp(
		o, nmDbusDest+".Device.Wireless",
		"ActiveAccessPoint", &d.accessPoint.objectPath,
	)
	if d.accessPoint.objectPath != "" {
		if err = d.accessPoint.loadProps(conn); err != nil {
			return
		}
	}
	return
}

type NmIp4Config struct {
	objectPath dbus.ObjectPath
	ip         net.IP
}

func (c *NmIp4Config) ipString() string {
	b, err := c.ip.MarshalText()
	if err != nil {
		return ""
	}
	return string(b)
}

func (c *NmIp4Config) loadProps(conn *dbus.Conn) (err error) {
	o := conn.Object(nmDbusDest, c.objectPath)
	var ipv4Addresses [][]uint32
	if err = dbusGetProp(o, nmDbusDest+".IP4Config", "Addresses", &ipv4Addresses); err != nil {
		return
	}
	c.ip = make(net.IP, 4)
	binary.BigEndian.PutUint32(c.ip, ipv4Addresses[0][0])
	return
}

type NmAccessPoint struct {
	objectPath dbus.ObjectPath
	strength   int
	ssid       string
}

func (a *NmAccessPoint) loadProps(conn *dbus.Conn) (err error) {
	o := conn.Object(nmDbusDest, a.objectPath)
	iface := nmDbusDest + ".AccessPoint"
	if err = dbusGetProp(o, iface, "Strength", &a.strength); err != nil {
		return
	}
	var ssidByte []byte
	if err = dbusGetProp(o, iface, "Ssid", &ssidByte); err != nil {
		return
	}
	a.ssid = string(ssidByte)
	return
}

type NetworkManagerBlock struct {
	NetworkManagerBlockConfig
	dbus                   *dbus.Conn
	connections            []NmActiveConnection
	currentConnectionIndex int
	state                  nmState
}

func NewNetworkManagerBlock() I3barBlocklet {
	b := NetworkManagerBlock{}
	b.Format = NewConfigFormatFromString("{state_icon$}{percentage*%}")
	b.AccessPointFormat = NewConfigFormatFromString("{strength:3*%$}{ssid}")
	b.currentConnectionIndex = 0
	b.PrimaryOnly = true
	return &b
}

func (t *NetworkManagerBlock) GetConfig() interface{} {
	return &t.NetworkManagerBlockConfig
}

func dbusGetProp(
	obj dbus.BusObject,
	iface, propName string,
	out interface{},
) error {
	return obj.Call(dbusGetProperty, 0, iface, propName).Store(out)
}

func (t *NetworkManagerBlock) loadConnections() (err error) {
	o := t.dbus.Object(nmDbusDest, nmDbusBasePath)
	t.connections = []NmActiveConnection{}
	if err = dbusGetProp(o, nmDbusDest, "State", &t.state); err != nil {
		return
	}
	var connPaths []dbus.ObjectPath
	if t.PrimaryOnly {
		var p dbus.ObjectPath
		if err = dbusGetProp(o, nmDbusDest, "PrimaryConnection", &p); err != nil {
			return
		}
		if p == "/" {
			return
		}
		connPaths = []dbus.ObjectPath{p}
	} else {
		if err = dbusGetProp(o, nmDbusDest, "ActiveConnections", &connPaths); err != nil {
			return
		}
	}
	t.connections = make([]NmActiveConnection, len(connPaths))
	for i := range t.connections {
		t.connections[i].objectPath = connPaths[i]
		if err := t.connections[i].loadProps(t.dbus); err != nil {
			return err
		}
	}
	return
}

func (t *NetworkManagerBlock) Run(ch UpdateChan, ctx context.Context) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		Log.Println("ConnectSystemBus", err)
		return
	}
	defer conn.Close()
	t.dbus = conn
	if err = conn.AddMatchSignal(
		dbus.WithMatchPathNamespace(nmDbusBasePath),
		dbus.WithMatchInterface(dbusPropertiesIface),
		dbus.WithMatchMember("PropertiesChanged"),
	); err != nil {
		Log.Print(err)
		return
	}
	if err := t.loadConnections(); err != nil {
		// Don't fail
		Log.Print(err)
	}
	c := make(chan *dbus.Signal)
	conn.Signal(c)
	ch.SendUpdate()
	for {
		select {
		case <-ctx.Done():
			return
		case sig := <-c:
			if len(t.connections) == 0 {
				t.loadConnections()
				ch.SendUpdate()
				break
			}
			changedProps := sig.Body[1].(map[string]dbus.Variant)
			shouldUpdate := false
			for i := range t.connections {
				conn := &t.connections[i]
				switch sig.Path {
				case nmDbusBasePath:
					if _, ok := changedProps["State"]; ok {
						t.loadConnections()
						ch.SendUpdate()
					}
					if _, ok := changedProps["PrimaryConnection"]; ok {
						if t.PrimaryOnly {
							t.loadConnections()
							ch.SendUpdate()
						}
					}
					if _, ok := changedProps["ActiveConnections"]; ok {
						if !t.PrimaryOnly {
							t.loadConnections()
							ch.SendUpdate()
						}
					}
				case conn.objectPath:
					conn.loadProps(t.dbus)
					ch.SendUpdate()
				case conn.device.objectPath:
					if st, ok := changedProps["Bitrate"]; ok {
						st.Store(&conn.device.bitrate)
						shouldUpdate = true
					}
				case conn.device.accessPoint.objectPath:
					if st, ok := changedProps["Strength"]; ok {
						st.Store(&conn.device.accessPoint.strength)
						shouldUpdate = true
					}
					if st, ok := changedProps["Ssid"]; ok {
						st.Store(&conn.device.accessPoint.ssid)
						shouldUpdate = true
					}
					// TODO: other properties
				case conn.ip4Config.objectPath:
					if st, ok := changedProps["Addresses"]; ok {
						var ipv4Addresses [][]uint32
						st.Store(&ipv4Addresses)
						binary.BigEndian.PutUint32(
							conn.ip4Config.ip,
							ipv4Addresses[0][0],
						)
					}
				}
			}
			if shouldUpdate {
				ch.SendUpdate()
			}
		}
	}
}

func (b *NetworkManagerBlock) Render(cfg *AppConfig) []I3barBlock {
	var iconName string
	var icon string
	switch b.state {
	case NM_STATE_UNKNOWN, NM_STATE_ASLEEP, NM_STATE_DISCONNECTED:
		iconName = "disconnected"
	case NM_STATE_CONNECTING, NM_STATE_DISCONNECTING:
		iconName = "connecting"
	case NM_STATE_CONNECTED_LOCAL, NM_STATE_CONNECTED_SITE:
		iconName = "connected_local"
	case NM_STATE_CONNECTED_GLOBAL:
		iconName = "connected"
	}
	if len(b.connections) == 0 {
		if i, ok := b.Icons[iconName].(string); ok {
			icon = i
		} else {
			for _, v := range b.Icons {
				if deviceStates, ok := v.(map[string]interface{}); ok {
					if i, ok := deviceStates[iconName].(string); ok {
						icon = i
						break
					}
				}
			}
		}
		return []I3barBlock{{
			FullText: b.Format.Expand(formatting.NamedArgs{
				"status_icon": icon,
			}),
		}}
	}
	c := b.connections[b.currentConnectionIndex]
	var iconMap map[string]interface{}
	switch c.device.deviceType {
	case NM_DEVICE_TYPE_ETHERNET:
		if m, ok := b.Icons["ethernet"].(map[string]interface{}); ok {
			iconMap = m
		}
	case NM_DEVICE_TYPE_WIFI:
		if m, ok := b.Icons["wifi"].(map[string]interface{}); ok {
			iconMap = m
		}
	default:
		return nil
	}
	if i, ok := iconMap[iconName].(string); ok {
		icon = i
	}
	accessPoint := ""
	switch b.state {
	case NM_STATE_CONNECTED_LOCAL,
		NM_STATE_CONNECTED_SITE,
		NM_STATE_CONNECTED_GLOBAL:
		if c.isWireless() {
			ap := c.device.accessPoint
			accessPoint = b.AccessPointFormat.Expand(formatting.NamedArgs{
				"strength": ap.strength,
				"ssid":     ap.ssid,
			})
			// TODO: refactor color applying
			icon = fmt.Sprintf(
				`<span color="%v">%s</span>`,
				cfg.Theme.HSVColor(PercentageToHue(ap.strength)),
				icon,
			)
		}
	}
	args := formatting.NamedArgs{
		"status_icon":  icon,
		"ipv4":         c.ip4Config.ipString(),
		"access_point": accessPoint,
	}
	if c.isVpn {
		if vpnIcon, ok := b.Icons["vpn"].(string); ok {
			args["vpn"] = vpnIcon
		}
	}
	return []I3barBlock{{
		FullText: b.Format.Expand(args),
		Markup:   MarkupPango,
	}}
}

func init() {
	RegisterBlocklet("networkmanager", NewNetworkManagerBlock)
}
