package blocks

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

type ShellBlockConfig struct {
	Command        string
	OnClickCommand string
	ListenClicks   bool
}

type ShellBlock struct {
	ShellBlockConfig
	lastText string
	ch       UpdateChan
}

func NewShellBlock(config *ShellBlockConfig) I3barBlocklet {
	block := ShellBlock{}
	block.Command = config.Command
	block.ch = make(chan int)
	return &block
}

// func (t *ShellBlock) runCommand() {
// 	cmd := exec.Command(t.Command[0], t.Command[1:]...)
// 	cmd.Env = append(
// 		os.Environ(),
// 	)
// }

func (t *ShellBlock) UpdateChan() UpdateChan {
	return t.ch
}

func (t *ShellBlock) Run() {
	cmd := exec.Command("sh", "-c", t.Command)
	rc, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	sc := bufio.NewScanner(rc)
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	for sc.Scan() {
		t.lastText = sc.Text()
		t.ch.SendUpdate()
	}
	if err := sc.Err(); err != nil {
		panic(err)
	}
}

func (t *ShellBlock) OnEvent(e I3barClickEvent) {
	cmd := exec.Command("sh", "-c", t.OnClickCommand)
	cmd.Env = append(
		os.Environ(),
		fmt.Sprintf("button=%d", e.Button),
		fmt.Sprintf("x=%d", e.Button),
		fmt.Sprintf("y=%d", e.Button),
	)
	if err := cmd.Run(); err != nil {
		panic(err)
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