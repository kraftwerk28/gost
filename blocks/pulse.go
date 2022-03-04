package blocks

import (
	"math"
	"time"

	. "github.com/kraftwerk28/gost/core"
	"github.com/kraftwerk28/gost/core/formatting"
	"github.com/lawl/pulseaudio"
)

type PulseConfig struct {
	SinkFormat   string `yaml:"sink_format"`
	SourceFormat string `yaml:"source_format"`
}

type NodeInfo struct {
	Volume uint32
	Icon   string
}

type PulseBlock struct {
	PulseConfig
	client       *pulseaudio.Client
	sink, source NodeInfo
}

func NewPulseBlock() I3barBlocklet {
	return &PulseBlock{}
}

func volumeToPercentage(v uint32) uint32 {
	return uint32(math.Round(float64(v) / 0xffff * 100))
}

func (c *PulseBlock) fetchInfo() {
	srv, _ := c.client.ServerInfo()
	sinks, _ := c.client.Sinks()
	for _, sink := range sinks {
		if sink.Name == srv.DefaultSink {
			c.sink.Volume = volumeToPercentage(sink.Cvolume[0])
			if sink.Muted {
				c.sink.Icon = "ﱝ "
				break
			}
			for _, port := range sink.Ports {
				if port.Name == sink.ActivePortName {
					switch port.Description {
					case "Speakers":
						c.sink.Icon = "墳"
					case "Headphones":
						c.sink.Icon = " "
					case "Headset":
						c.sink.Icon = " "
					}
				}
			}
			break
		}
	}
	sources, _ := c.client.Sources()
	for _, source := range sources {
		if source.Name == srv.DefaultSource {
			c.source.Volume = volumeToPercentage(source.Cvolume[0])
			if source.Muted {
				c.source.Icon = " "
				break
			}
			for _, port := range source.Ports {
				if port.Name == source.ActivePortName {
					c.source.Icon = ""
				}
			}
			break
		}
	}
}

const throttleDuration = time.Millisecond * 50

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

func (c *PulseBlock) Run(ch UpdateChan) {
	client, _ := pulseaudio.NewClient()
	c.client = client
	upd, _ := client.Updates()
	c.fetchInfo()
	throttleTimer := time.NewTimer(throttleDuration)
	for {
		select {
		case <-upd:
			throttleTimer.Reset(throttleDuration)
		case <-throttleTimer.C:
			c.fetchInfo()
			ch.SendUpdate()
		}
	}
}

func (c *PulseBlock) GetConfig() interface{} {
	return &c.PulseConfig
}

func (t *PulseBlock) Render() []I3barBlock {
	f := formatting.GoFmt{}
	return []I3barBlock{
		{
			FullText: f.Sprintf(t.SinkFormat, formatting.NamedArgs{
				"sink_icon":   t.sink.Icon,
				"sink_volume": t.sink.Volume,
			}),
			Name: "sink",
		},
		{
			FullText: f.Sprintf(t.SourceFormat, formatting.NamedArgs{
				"source_icon":   t.source.Icon,
				"source_volume": t.source.Volume,
			}),
			Name: "source",
		},
	}
}

func (t *PulseBlock) OnEvent(e *I3barClickEvent) {
	n := e.CustomBlockletName()
	switch n {
	case "sink":
		switch e.Button {
		case ButtonScrollDown:
			s, _ := t.getCurrentSink()
			t.client.SetSinkVolume(s.Name, float32(s.Cvolume[0])/0xffff+0.01)
		case ButtonScrollUp:
			s, _ := t.getCurrentSink()
			t.client.SetSinkVolume(s.Name, float32(s.Cvolume[0])/0xffff-0.01)
		}
	case "source":
		// TODO: change mic volume
	}
}

func init() {
	RegisterBlocklet("pulseaudio", NewPulseBlock)
}
