package blocks

import (
	"bufio"
	"bytes"
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

type ShellCmd struct {
	*exec.Cmd
	stdout, stderr bytes.Buffer
}

func (t *ShellBlock) newCmd(ctx context.Context) *exec.Cmd {
	return exec.CommandContext(ctx, "sh", "-c", t.Command)
}

func (s *ShellCmd) Exec() (stdout, stderr []byte, exitErr *exec.ExitError) {
	err := s.Run()
	if err != nil {
		if casted, ok := err.(*exec.ExitError); ok {
			return nil, s.stderr.Bytes(), casted
		}
		return nil, nil, nil
	}
	return s.stdout.Bytes(), nil, nil
}

func processCmdOutput(o []byte) string {
	return strings.TrimSpace(string(o))
}

func (b *ShellBlock) Run(ch UpdateChan, ctx context.Context) {
	// The blocklet runs in 3 modes:
	if b.Interval != nil {
		// interval: run script, display output, sleep, repeat
		t := time.NewTicker(time.Duration(*b.Interval))
		for {
			stdout := bytes.Buffer{}
			cmd := b.newCmd(ctx)
			cmd.Stdout, cmd.Stderr = &stdout, Log.Writer()
			err := cmd.Run()
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.Exited() {
				panic(exitErr)
			} else {
				b.lastText = processCmdOutput(stdout.Bytes())
				ch.SendUpdate()
			}
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	} else if b.RestartOnExit {
		// restart_on_exit: run script, wait for exit, display output, repeat
		t := time.NewTimer(time.Second)
		for {
			t.Reset(time.Second)
			stdout := bytes.Buffer{}
			cmd := b.newCmd(ctx)
			cmd.Stdout, cmd.Stderr = &stdout, Log.Writer()
			err := cmd.Run()
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.Exited() {
				panic(exitErr)
			} else {
				b.lastText = processCmdOutput(stdout.Bytes())
				ch.SendUpdate()
			}
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				// Prevent spamming
			}
		}
	} else {
		// continous: run script, read lines from stdout, each line updates the text
		cmd := b.newCmd(ctx)
		rc, err := cmd.StdoutPipe()
		cmd.Stderr = Log.Writer()
		if err != nil {
			panic(err)
		}
		defer rc.Close()
		if err := cmd.Start(); err != nil {
			panic(err)
		}
		sc := bufio.NewScanner(rc)
		for sc.Scan() {
			b.lastText = processCmdOutput(sc.Bytes())
			ch.SendUpdate()
		}
		if err = sc.Err(); err != nil {
			panic(sc.Err())
		}
		err = cmd.Wait()
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.Exited() {
			panic(err)
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
