package blocks

type nmState uint32

const (
	// Networking state is unknown. This indicates a daemon error that makes it
	// unable to reasonably assess the state. In such event the applications
	// are expected to assume Internet connectivity might be present and not
	// disable controls that require network access. The graphical shells may
	// hide the network accessibility indicator altogether since no meaningful
	// status indication can be provided.
	NM_STATE_UNKNOWN nmState = 0
	// Networking is not enabled, the system is being suspended or resumed from
	// suspend.
	NM_STATE_ASLEEP = 10
	// There is no active network connection. The graphical shell should
	// indicate no network connectivity and the applications should not attempt
	// to access the network.
	NM_STATE_DISCONNECTED = 20
	// Network connections are being cleaned up. The applications should tear
	// down their network sessions.
	NM_STATE_DISCONNECTING = 30
	// A network connection is being started The graphical shell should
	// indicate the network is being connected while the applications should
	// still make no attempts to connect the network.
	NM_STATE_CONNECTING = 40
	// There is only local IPv4 and/or IPv6 connectivity, but no default route
	// to access the Internet. The graphical shell should indicate no network
	// connectivity.
	NM_STATE_CONNECTED_LOCAL = 50
	// There is only site-wide IPv4 and/or IPv6 connectivity. This means a
	// default route is available, but the Internet connectivity check (see
	// "Connectivity" property) did not succeed. The graphical shell should
	// indicate limited network connectivity.
	NM_STATE_CONNECTED_SITE = 60
	// There is global IPv4 and/or IPv6 Internet connectivity This means the
	// Internet connectivity check succeeded, the graphical shell should
	// indicate full network connectivity.
	NM_STATE_CONNECTED_GLOBAL = 70
)

type nmActiveConnectionState uint32

const (
	// the state of the connection is unknown
	NM_ACTIVE_CONNECTION_STATE_UNKNOWN nmActiveConnectionState = 0
	// a network connection is being prepared
	NM_ACTIVE_CONNECTION_STATE_ACTIVATING = 1
	// there is a connection to the network
	NM_ACTIVE_CONNECTION_STATE_ACTIVATED = 2
	// the network connection is being torn down and cleaned up
	NM_ACTIVE_CONNECTION_STATE_DEACTIVATING = 3
	// the network connection is disconnected and will be removed
	NM_ACTIVE_CONNECTION_STATE_DEACTIVATED = 4
)

type nmDeviceType uint32

const (
	// unknown device
	NM_DEVICE_TYPE_UNKNOWN nmDeviceType = 0
	// generic support for unrecognized device types
	NM_DEVICE_TYPE_GENERIC = 14
	// a wired ethernet device
	NM_DEVICE_TYPE_ETHERNET = 1
	// an 802.11 Wi-Fi device
	NM_DEVICE_TYPE_WIFI = 2
	// not used
	NM_DEVICE_TYPE_UNUSED1 = 3
	// not used
	NM_DEVICE_TYPE_UNUSED2 = 4
	// a Bluetooth device supporting PAN or DUN access protocols
	NM_DEVICE_TYPE_BT = 5
	// an OLPC XO mesh networking device
	NM_DEVICE_TYPE_OLPC_MESH = 6
	// an 802.16e Mobile WiMAX broadband device
	NM_DEVICE_TYPE_WIMAX = 7
	// a modem supporting analog telephone, CDMA/EVDO, GSM/UMTS, or LTE network
	// access protocols
	NM_DEVICE_TYPE_MODEM = 8
	// an IP-over-InfiniBand device
	NM_DEVICE_TYPE_INFINIBAND = 9
	// a bond master interface
	NM_DEVICE_TYPE_BOND = 10
	// an 802.1Q VLAN interface
	NM_DEVICE_TYPE_VLAN = 11
	// ADSL modem
	NM_DEVICE_TYPE_ADSL = 12
	// a bridge master interface
	NM_DEVICE_TYPE_BRIDGE = 13
	// a team master interface
	NM_DEVICE_TYPE_TEAM = 15
	// a TUN or TAP interface
	NM_DEVICE_TYPE_TUN = 16
	// a IP tunnel interface
	NM_DEVICE_TYPE_IP_TUNNEL = 17
	// a MACVLAN interface
	NM_DEVICE_TYPE_MACVLAN = 18
	// a VXLAN interface
	NM_DEVICE_TYPE_VXLAN = 19
	// a VETH interface
	NM_DEVICE_TYPE_VETH = 20
	// a MACsec interface
	NM_DEVICE_TYPE_MACSEC = 21
	// a dummy interface
	NM_DEVICE_TYPE_DUMMY = 22
	// a PPP interface
	NM_DEVICE_TYPE_PPP = 23
	// a Open vSwitch interface
	NM_DEVICE_TYPE_OVS_INTERFACE = 24
	// a Open vSwitch port
	NM_DEVICE_TYPE_OVS_PORT = 25
	// a Open vSwitch bridge
	NM_DEVICE_TYPE_OVS_BRIDGE = 26
	// a IEEE 802.15.4 (WPAN) MAC Layer Device
	NM_DEVICE_TYPE_WPAN = 27
	// 6LoWPAN interface
	NM_DEVICE_TYPE_6LOWPAN = 28
	// a WireGuard interface
	NM_DEVICE_TYPE_WIREGUARD = 29
	// an 802.11 Wi-Fi P2P device. Since: 1.16.
	NM_DEVICE_TYPE_WIFI_P2P = 30
	// A VRF (Virtual Routing and Forwarding) interface. Since: 1.24.
	NM_DEVICE_TYPE_VRF = 31
)

type nmDeviceState uint32

const (
	// the device's state is unknown
	NM_DEVICE_STATE_UNKNOWN nmDeviceState = 0
	// the device is recognized, but not managed by NetworkManager
	NM_DEVICE_STATE_UNMANAGED = 10
	// the device is managed by NetworkManager, but is not available for use.
	// Reasons may include the wireless switched off, missing firmware, no
	// ethernet carrier, missing supplicant or modem manager, etc.
	NM_DEVICE_STATE_UNAVAILABLE = 20
	// the device can be activated, but is currently idle and not connected to
	// a network.
	NM_DEVICE_STATE_DISCONNECTED = 30
	// the device is preparing the connection to the network. This may include
	// operations like changing the MAC address, setting physical link
	// properties, and anything else required to connect to the requested
	// network.
	NM_DEVICE_STATE_PREPARE = 40
	// the device is connecting to the requested network. This may include
	// operations like associating with the Wi-Fi AP, dialing the modem,
	// connecting to the remote Bluetooth device, etc.
	NM_DEVICE_STATE_CONFIG = 50
	// the device requires more information to continue connecting to the
	// requested network. This includes secrets like WiFi passphrases, login
	// passwords, PIN codes, etc.
	NM_DEVICE_STATE_NEED_AUTH = 60
	// the device is requesting IPv4 and/or IPv6 addresses and routing
	// information from the network.
	NM_DEVICE_STATE_IP_CONFIG = 70
	// the device is checking whether further action is required for the
	// requested network connection. This may include checking whether only
	// local network access is available, whether a captive portal is blocking
	// access to the Internet, etc.
	NM_DEVICE_STATE_IP_CHECK = 80
	// the device is waiting for a secondary connection (like a VPN) which must
	// activated before the device can be activated
	NM_DEVICE_STATE_SECONDARIES = 90
	// the device has a network connection, either local or global.
	NM_DEVICE_STATE_ACTIVATED = 100
	// a disconnection from the current network connection was requested, and
	// the device is cleaning up resources used for that connection. The
	// network connection may still be valid.
	NM_DEVICE_STATE_DEACTIVATING = 110
	// the device failed to connect to the requested network and is cleaning up
	// the connection request
	NM_DEVICE_STATE_FAILED = 120
)
