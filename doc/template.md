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
$ curl -L -o "${XDG_CONFIG_HOME:-$HOME/.config}/gost/config.yml" \
    https://raw.githubusercontent.com/kraftwerk28/gost/master/doc/example-config.yml
```


## Blocklets

{{range .Blocklets -}}
#### [`{{.Name}}`]({{.FileName}})

{{if .Doc -}}
{{.Doc}}
{{end -}}
{{if .ConfigFields -}}
| Option | Type | Description |
|---|---|---|
{{range .ConfigFields -}}
| `{{.Name}}` | {{if .DataType}}`{{.DataType}}`{{end}} | {{.Doc | escapeTableDelimiters}} |
{{end -}}
{{else -}}
_No config fields._
{{end}}

{{end -}}


---

This project is inspired by
[i3status-rust](https://github.com/greshake/i3status-rust) and
[i3blocks](https://github.com/vivien/i3blocks)
