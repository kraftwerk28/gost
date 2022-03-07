package blocks

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"

	. "github.com/kraftwerk28/gost/core"
)

type ShellBlockConfig struct {
	Command        string `yaml:"command"`
	OnClickCommand string `yaml:"on_click"`
}

type ShellBlock struct {
	*ShellBlockConfig
	lastText string
}

func NewShellBlock() I3barBlocklet {
	return &ShellBlock{ShellBlockConfig: new(ShellBlockConfig)}
}

func (s *ShellBlock) GetConfig() interface{} {
	return &s.ShellBlockConfig
}

func (t *ShellBlock) Run(ch UpdateChan, ctx context.Context) {
	cmd := exec.Command("sh", "-c", t.Command)
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
		t.lastText = sc.Text()
		ch.SendUpdate()
	}
	if err := sc.Err(); err != nil {
		Log.Println(err)
	}
}

func (t *ShellBlock) OnEvent(e *I3barClickEvent) {
	if t.OnClickCommand == "" {
		return
	}
	cmd := exec.Command("sh", "-c", t.OnClickCommand)
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("button=%d", e.Button),
		fmt.Sprintf("x=%d", e.Button),
		fmt.Sprintf("y=%d", e.Button),
	)
	cmd.Stdout = Log.Writer()
	if err := cmd.Run(); err != nil {
		Log.Println(err)
	} else {
		Log.Println("OnClickCommand exited.")
	}
}

func (t *ShellBlock) Render() []I3barBlock {
	if t.lastText == "" {
		return nil
	}
	return []I3barBlock{{
		FullText: t.lastText,
		// TODO: auto-assigned name
		Name: "shell",
	}}
}

func init() {
	RegisterBlocklet("shell", NewShellBlock)
}
