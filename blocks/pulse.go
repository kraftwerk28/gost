package blocks

import (
	"context"
	"fmt"
	"math"
	"time"

	. "github.com/kraftwerk28/gost/core"
	"github.com/kraftwerk28/gost/core/formatting"
	"github.com/kraftwerk28/pulseaudio"
)

const (
	nodeKindSink   = "sink"
	nodeKindSource = "source"
)

type PulseIconsConfig struct {
	Devices     map[string]string `yaml:"devices"`
	SinkMuted   string            `yaml:"sink_muted"`
	SourceMuted string            `yaml:"source_muted"`
}

type PulseConfig struct {
	BaseBlockletConfig `yaml:",inline"`
	Node               string           `yaml:"node"`
	Format             *ConfigFormat    `yaml:"format"`
	Icons              PulseIconsConfig `yaml:"icons"`
}

type PulseBlock struct {
	PulseConfig
	client   *pulseaudio.Client
	volume   uint32
	muted    bool
	portDesc string
	// icon   string
}

func NewPulseBlock() I3barBlocklet {
	return &PulseBlock{}
}

func volumeToPercentage(v uint32) uint32 {
	ratio := float64(v) / 0xffff
	return uint32(math.Round(ratio * 100))
}

// Updates from pulse server are bursting, so some throttling is required
const throttleDuration = time.Millisecond * 50

func (c *PulseBlock) fetchInfo() bool {
	switch c.Node {
	case nodeKindSource:
		source, err := c.getCurrentSource()
		if source == nil {
			return false
		}
		if err != nil {
			Log.Print(err)
		}
		c.volume = volumeToPercentage(source.Cvolume[0])
		if source.Muted {
			c.muted = true
			break
		}
		c.muted = false
		for _, port := range source.Ports {
			if port.Name == source.ActivePortName {
				c.portDesc = port.Description
				break
			}
		}
	case nodeKindSink:
		sink, err := c.getCurrentSink()
		if sink == nil {
			return false
		}
		if err != nil {
			Log.Print(err)
		}
		c.volume = volumeToPercentage(sink.Cvolume[0])
		if sink.Muted {
			c.muted = true
			break
		}
		c.muted = false
		for _, port := range sink.Ports {
			if port.Name == sink.ActivePortName {
				c.portDesc = port.Description
				break
			}
		}
	}
	return true
}

func (c *PulseBlock) getCurrentSink() (*pulseaudio.Sink, error) {
	srv, err := c.client.ServerInfo()
	if err != nil {
		return nil, err
	}
	sinks, err := c.client.Sinks()
	if err != nil {
		return nil, err
	}
	for _, sink := range sinks {
		if sink.Name == srv.DefaultSink {
			return &sink, nil
		}
	}
	return nil, nil
}

func (c *PulseBlock) getCurrentSource() (*pulseaudio.Source, error) {
	srv, err := c.client.ServerInfo()
	if err != nil {
		return nil, err
	}
	sources, err := c.client.Sources()
	if err != nil {
		return nil, err
	}
	for _, source := range sources {
		if source.Name == srv.DefaultSource {
			return &source, nil
		}
	}
	return nil, nil
}

func (c *PulseBlock) Run(ch UpdateChan, ctx context.Context) {
	client, err := pulseaudio.NewClient()
	if err != nil {
		Log.Print(err)
		return
	}
	defer client.Close()
	c.client = client
	upd, err := client.Updates()
	if err != nil {
		Log.Print(err)
		return
	}
	c.fetchInfo()
	throttleTimer := time.NewTimer(throttleDuration)
	for {
		select {
		case <-upd:
			throttleTimer.Reset(throttleDuration)
		case <-throttleTimer.C:
			if c.fetchInfo() {
				ch.SendUpdate()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *PulseBlock) GetConfig() interface{} {
	return &c.PulseConfig
}

func (t *PulseBlock) Render(cfg *AppConfig) []I3barBlock {
	var icon string
	if t.muted {
		switch t.Node {
		case nodeKindSink:
			icon = t.Icons.SinkMuted
		case nodeKindSource:
			icon = t.Icons.SourceMuted
		}
		icon = fmt.Sprintf(
			`<span color="%v">%s</span>`,
			cfg.Theme.HSVColor(0),
			icon,
		)
	} else {
		icon = t.Icons.Devices[t.portDesc]
	}
	return []I3barBlock{{
		FullText: t.Format.Expand(formatting.NamedArgs{
			"icon":   icon,
			"volume": t.volume,
		}),
		Name:   nodeKindSink,
		Markup: MarkupPango,
	}}
}

func (t *PulseBlock) changeVolume(delta int) error {
	d := float32(delta) * 0.01
	switch t.Node {
	case nodeKindSink:
		s, err := t.getCurrentSink()
		if err != nil {
			return err
		}
		vol := float32(s.Cvolume[0]) / 0xffff
		vol += d
		return t.client.SetSinkVolume(s.Name, vol)
	case nodeKindSource:
		s, err := t.getCurrentSource()
		if err != nil {
			return err
		}
		vol := float32(s.Cvolume[0]) / 0xffff
		vol += d
		return t.client.SetSourceVolume(s.Name, vol)
	}
	return nil
}

func (t *PulseBlock) OnEvent(e *I3barClickEvent, ctx context.Context) {
	switch e.Button {
	case ButtonScrollDown:
		t.changeVolume(-1)
	case ButtonScrollUp:
		t.changeVolume(+1)
	}
}

func init() {
	RegisterBlocklet("pulseaudio", NewPulseBlock)
}
