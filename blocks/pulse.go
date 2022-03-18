package blocks

import (
	"context"
	"math"
	"time"

	. "github.com/kraftwerk28/gost/core"
	"github.com/kraftwerk28/gost/core/formatting"
	"github.com/lawl/pulseaudio"
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
	client *pulseaudio.Client
	Volume uint32
	Icon   string
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
	case "source":
		source, err := c.getCurrentSource()
		if source == nil {
			return false
		}
		if err != nil {
			Log.Print(err)
		}
		c.Volume = volumeToPercentage(source.Cvolume[0])
		if source.Muted {
			c.Icon = c.Icons.SourceMuted
			break
		}
		for _, port := range source.Ports {
			if port.Name == source.ActivePortName {
				c.Icon = c.Icons.Devices[port.Description]
				break
			}
		}
	case "sink":
		sink, err := c.getCurrentSink()
		if sink == nil {
			return false
		}
		if err != nil {
			Log.Print(err)
		}
		c.Volume = volumeToPercentage(sink.Cvolume[0])
		if sink.Muted {
			c.Icon = c.Icons.SinkMuted
			break
		}
		for _, port := range sink.Ports {
			if port.Name == sink.ActivePortName {
				c.Icon = c.Icons.Devices[port.Description]
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
	return []I3barBlock{
		{
			FullText: t.Format.Expand(formatting.NamedArgs{
				"icon":   t.Icon,
				"volume": t.Volume,
			}),
			Name: "sink",
		},
	}
}

func (t *PulseBlock) changeVolume(delta int) {
	d := float32(delta) * 0.01
	switch t.Node {
	case "sink":
		s, _ := t.getCurrentSink()
		vol := float32(s.Cvolume[0]) / 0xffff
		vol += d
		t.client.SetSinkVolume(s.Name, vol)
	case "source":
		// s, _ := t.getCurrentSource()
		// vol := float32(s.Cvolume[0]) / 0xffff
		Log.Println("Changing source volume it not supported")
	}
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
