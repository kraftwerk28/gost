package blocks

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

type ShellBlockConfig struct {
	Command      []string
	ListenClicks bool
}

type ShellBlock struct {
	ShellBlockConfig
	lastText string
}

func NewShellBlock(config *ShellBlockConfig) I3barBlocklet {
	block := ShellBlock{}
	block.Command = config.Command
	return &block
}

func (t *ShellBlock) runCommand() {
	cmd := exec.Command(t.Command[0], t.Command[1:]...)
	cmd.Env = append(
		os.Environ(),
	)
}

func (t *ShellBlock) Run() {
	if t.ListenClicks {
		for {
		}
	}
}

func (t *ShellBlock) OnEvent(e *I3barClickEvent) {
}

func (t *ShellBlock) Render() []I3barBlock {
	currentTime := time.Now()
	return []I3barBlock{{
		FullText: fmt.Sprintf(
			"%d.%d %d:%d:%d",
			currentTime.Day(),
			currentTime.Month(),
			currentTime.Hour(),
			currentTime.Minute(),
			currentTime.Second(),
		),
		// TODO: auto-assigned name
		Name: "myclock",
	}}
}
