# gost

### A JSON feeder for i3bar/swaybar


## Installation

##### TBD...


## Configuration

The application is configured via a file in
`${XDG_CONFIG_HOME}/gost/config.yml` (usually `$HOME/.config/gost/config.yml`).

_config.yml_:

```yaml
version: "1"

theme:
  hue: 0.7
  saturation: 0.7

separator_width: 16

blocks:
  - name: foo_block
    config_option_1: bar_value
    config_option_2:
      - baz_value
```

_sway/config_ (or _i3/config_):
```i3config
bar {
  status_command gost
  # ...
}
```

You can grab the example config from [doc/example-config.yml](doc/example-config.yml):

```bash
$ curl -sL -o "${XDG_CONFIG_HOME:-$HOME/.config}/gost/config.yml" \
    https://raw.githubusercontent.com/kraftwerk28/gost/master/doc/example-config.yml
```


## Blocklets

#### [`battery`](blocks/battery.go)

Display battery charge level. Requires UPower to work over DBus
`upower_device` can be obtained via DBus using the following command:
```bash
$ busctl --json=pretty --system call org.freedesktop.UPower \
  /org/freedesktop/UPower org.freedesktop.UPower \
  EnumerateDevices \
  | jq -r '.data[0] | map(match("[^/]+$").string)[]'`
```
If empty, the program will try to detect battery device.

| Option | Type | Description |
|---|---|---|
| `format` | `ConfigFormat` |  |
| `upower_device` | `string` | Device name. See above. |
| `urgent_level` | `int` |  |


#### [`bluez`](blocks/bluez.go)

Displays connected bluetooth devices

| Option | Type | Description |
|---|---|---|
| `mac` | `string` | Mac address of the device |
| `format` | `ConfigFormat` |  |
| `device_format` | `ConfigFormat` | Device format |


#### [`clickcount`](blocks/clickcount.go)

| Option | Type | Description |
|---|---|---|
| `format` | `ConfigFormat` |  |


#### [`dbus`](blocks/dbus.go)

| Option | Type | Description |
|---|---|---|
| `object_path` | `string` |  |
| `initial_text` | `string` |  |


#### [`lua`](blocks/lua.go)

| Option | Type | Description |
|---|---|---|
| `script` | `string` |  |


#### [`mpris`](blocks/mpris.go)

| Option | Type | Description |
|---|---|---|
| `player_format` | `ConfigFormat` |  |
| `separator` | `string` |  |


#### [`network_manager`](blocks/networkmanager.go)

| Option | Type | Description |
|---|---|---|
| `format` | `ConfigFormat` |  |
| `ap_format` | `ConfigFormat` |  |
| `primary_only` | `bool` |  |


#### [`pulse`](blocks/pulse.go)

| Option | Type | Description |
|---|---|---|
| `node` | `string` |  |
| `format` | `ConfigFormat` |  |
| `icons` | `PulseIconsConfig` |  |


#### [`shell`](blocks/shell.go)

Displays output for a shell script
It has three modes:
- run script every N seconds, wait for it to exit
- run script once, read lines from it's stdout
  and update the blocklet per each line
- run script in a loop as soon as it exits

| Option | Type | Description |
|---|---|---|
| `command` | `string` | Shell command to run |
| `on_click` | `string` | Shell command to run when a blocklet is clicked |
| `interval` | `ConfigInterval` |  |
| `restart_on_exit` | `bool` |  |
| `json` | `bool` | Treat command's output as json instead of plain text |


#### [`sway_layout`](blocks/sway_layout.go)

Displays current keyboard layout.
Uses sway's IPC API for retrieving the info (i.e. won't work with i3wm)

| Option | Type | Description |
|---|---|---|
| `format` | `ConfigFormat` |  |
| `input` | `string` |  |


#### [`sway_window`](blocks/sway_window.go)

_No config fields._


#### [`time`](blocks/time.go)

| Option | Type | Description |
|---|---|---|
| `format` | `ConfigFormat` |  |
| `layout` | `string` |  |
| `interval` | `ConfigInterval` |  |


---

This project is inspired by
[i3status-rust](https://github.com/greshake/i3status-rust) and
[i3blocks](https://github.com/vivien/i3blocks)
