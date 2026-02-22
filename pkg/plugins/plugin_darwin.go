package plugins

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"
	"text/template"

	"github.com/pkg/errors"
)

func Setpgid(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

func (p *Plugin) runInTerminal(appleScriptTemplate3, command, paramsStr string, vars []string) error {
	tpl, err := template.New("appleScriptTemplate3").Parse(appleScriptTemplate3)
	if err != nil {
		return err
	}
	commandLine := command
	var renderedScript bytes.Buffer
	err = tpl.Execute(&renderedScript, struct {
		Command string
		Vars    string
		Params  string
	}{
		Command: commandLine,
		Vars:    fmt.Sprintf("%q", variablesEnvString(vars)),
		Params:  paramsStr,
	})
	if err != nil {
		return err
	}
	appleScript := renderedScript.String()
	cmd := exec.Command("osascript", "-s", "h", "-e", appleScript)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
	}
	if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
		return errors.Errorf("run in terminal script failed: %s", stderr.String())
	}
	return nil
}

func runInTerminalDirect(appleScriptTemplate3, command, paramsStr string, vars []string) error {
	tpl, err := template.New("appleScriptTemplate3").Parse(appleScriptTemplate3)
	if err != nil {
		return err
	}
	var renderedScript bytes.Buffer
	err = tpl.Execute(&renderedScript, struct {
		Command string
		Vars    string
		Params  string
	}{
		Command: command,
		Vars:    fmt.Sprintf("%q", variablesEnvString(vars)),
		Params:  paramsStr,
	})
	if err != nil {
		return err
	}
	appleScript := renderedScript.String()
	cmd := exec.Command("osascript", "-s", "h", "-e", appleScript)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
		return errors.Errorf("run in terminal script failed: %s", stderr.String())
	}
	return nil
}
