package blocks

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os/exec"
	"strings"
	"time"

	. "github.com/kraftwerk28/gost/core"
)

type ShellBlockConfig struct {
	Command        string          `yaml:"command"`
	OnClickCommand *string         `yaml:"on_click"`
	Interval       *ConfigInterval `yaml:"interval"`
	RestartOnExit  bool            `yaml:"restart_on_exit"`
	Json           bool            `yaml:"json"`
}

type ShellBlock struct {
	*ShellBlockConfig
	lastText string
	log      *log.Logger
}

func NewShellBlock() I3barBlocklet {
	return &ShellBlock{ShellBlockConfig: new(ShellBlockConfig)}
}

func (s *ShellBlock) GetConfig() interface{} {
	return &s.ShellBlockConfig
}

func (t *ShellBlock) newCmd(ctx context.Context) *exec.Cmd {
	return exec.CommandContext(ctx, "sh", "-c", t.Command)
}

func processCmdOutput(o []byte) string {
	return strings.TrimSpace(string(o))
}

func (t *ShellBlock) Run(ch UpdateChan, ctx context.Context) {
	if t.RestartOnExit {
		for {
			cmd := t.newCmd(ctx)
			outp, err := cmd.Output()
			if err != nil {
				Log.Print(err)
			} else {
				t.lastText = processCmdOutput(outp)
				ch.SendUpdate()
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	} else if t.Interval != nil {
		ticker := time.Tick(time.Duration(*t.Interval))
	loop:
		for {
			cmd := t.newCmd(ctx)
			outp, err := cmd.Output()
			if e, ok := err.(*exec.ExitError); ok {
				Log.Printf(
					"Command `%s` errored: %s\n%s",
					t.Command, e, e.Stderr,
				)
			} else {
				t.lastText = processCmdOutput(outp)
				ch.SendUpdate()
			}
			select {
			case <-ticker:
				continue loop
			case <-ctx.Done():
				break loop
			}
		}
	} else {
		cmd := t.newCmd(ctx)
		rc, err := cmd.StdoutPipe()
		defer rc.Close()
		if err != nil {
			// TODO: per-block logger
			Log.Print(err)
		}
		sc := bufio.NewScanner(rc)
		if err := cmd.Start(); err != nil {
			Log.Print(err)
		}
		for sc.Scan() {
			t.lastText = processCmdOutput(sc.Bytes())
			ch.SendUpdate()
		}
		if err := sc.Err(); err != nil {
			Log.Println(err)
		}
	}
}

func (t *ShellBlock) OnEvent(e *I3barClickEvent, ctx context.Context) {
	if t.OnClickCommand == nil {
		return
	}
	cmd := e.ShellCommand(*t.OnClickCommand, ctx)
	if err := cmd.Run(); err != nil {
		Log.Println(err)
	}
}

func (t *ShellBlock) Render() []I3barBlock {
	if t.lastText == "" {
		return nil
	}
	if t.Json {
		var block I3barBlock
		if err := json.Unmarshal([]byte(t.lastText), &block); err != nil {
			Log.Print(err)
			return nil
		}
		return []I3barBlock{block}
	} else {
		return []I3barBlock{{FullText: t.lastText}}
	}
}

func init() {
	RegisterBlocklet("shell", NewShellBlock)
}
