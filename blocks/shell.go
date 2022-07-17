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

type ShellBlock struct {
	ShellBlockConfig
	lastText string
	cmd      *exec.Cmd
	log      *log.Logger
}

// Displays output for a shell script
// It has three modes:
// - run script every N seconds, wait for it to exit
// - run script once, read lines from it's stdout
//   and update the blocklet per each line
// - run script in a loop as soon as it exits
type ShellBlockConfig struct {
	// Shell command to run
	Command string `yaml:"command"`
	// Shell command to run when a blocklet is clicked
	OnClickCommand *string         `yaml:"on_click"`
	Interval       *ConfigInterval `yaml:"interval"`
	RestartOnExit  bool            `yaml:"restart_on_exit"`
	// Treat command's output as json instead of plain text
	Json bool `yaml:"json"`
}

func NewShellBlock() I3barBlocklet {
	return &ShellBlock{}
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
	time.Sleep(time.Second)

	// go func() {
	// 	<- ctx.Done()
	// 	if t.cmd != nil {
	// 		t.cmd.Process.Kill()
	// 	}
	// }()
	if t.RestartOnExit {
		for {
			cmd := t.newCmd(ctx)
			outp, err := cmd.Output()
			if err, ok := err.(*exec.ExitError); ok {
			} else if err != nil {
			} else {
			}

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
		tickTimer := time.NewTicker(time.Duration(*t.Interval))
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
			case <-tickTimer.C:
				continue loop
			case <-ctx.Done():
				break loop
			}
		}
	} else {
		cmd := t.newCmd(ctx)
		rc, err := cmd.StdoutPipe()
		if err != nil {
			// TODO: per-block logger
			Log.Print(err)
		}
		defer rc.Close()
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

func (t *ShellBlock) Render(cfg *AppConfig) []I3barBlock {
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
