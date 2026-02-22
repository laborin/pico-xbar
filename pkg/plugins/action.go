package plugins

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ActionFunc is a function that handles the
// menu item clicks/selections.
type ActionFunc func(ctx context.Context)

// actionTimeout is the amount of time xbar will wait for an
// action to complete.
const actionTimeout = 10 * time.Second

// Action returns a function that will handle the
// action should this item be clicked/selected.
// nil response indicates no action, so you must check
// for nil before calling.
// The following code should be called:
//
//	actionFunc := item.Action()
//	if actionFunc != nil {
//		actionFunc(ctx)
//	}
func (i *Item) Action() ActionFunc {
	var plugin *Plugin
	if i.Plugin != nil {
		plugin = i.Plugin
	}
	var actions []ActionFunc
	if i.Params.Href != "" {
		actions = append(actions, actionHref(i.Params.Href))
	}
	if i.Params.Shell != "" {
		// Capture only needed values to avoid keeping entire Item alive
		pluginDir := ""
		appleScriptTemplate := ""
		var envVars []string
		if plugin != nil {
			pluginDir = filepath.Dir(plugin.Command)
			appleScriptTemplate = plugin.AppleScriptTemplate
			envVars = plugin.Variables
		}
		actions = append(actions, actionShellWithDir(pluginDir, i.Params.Terminal, appleScriptTemplate, i.Params.Shell, i.Params.ShellParams, envVars))
	}
	if i.Params.Refresh && plugin != nil {
		shouldDelayBeforeRefresh := len(actions) > 0
		// Capture plugin pointer directly to avoid keeping Item alive
		actions = append(actions, actionRefresh(func(ctx context.Context) {
			if shouldDelayBeforeRefresh {
				time.Sleep(500 * time.Millisecond)
			}
			plugin.TriggerRefresh()
		}))
	}
	if len(actions) == 0 {
		return nil // no actions
	}
	return actionFuncs(actions...)
}

// actionFuncs makes an ActionFunc that runs multuple functions
// in order.
func actionFuncs(actions ...ActionFunc) ActionFunc {
	return func(ctx context.Context) {
		for i := range actions {
			if err := ctx.Err(); err != nil {
				return // don't bother - context cancelled
			}
			fn := actions[i]
			fn(ctx)
		}
	}
}

// actionHref gets an ActionFunc that opens a URL.
func actionHref(href string) ActionFunc {
	return func(ctx context.Context) {
		commandCtx, cancel := context.WithTimeout(ctx, actionTimeout)
		defer cancel()
		switch runtime.GOOS {
		case "linux":
			cmd := exec.CommandContext(commandCtx, "xdg-open", href)
			Setpgid(cmd)
			_ = cmd.Run()
		case "windows":
			cmd := exec.CommandContext(commandCtx, "rundll32", "url.dll,FileProtocolHandler", href)
			Setpgid(cmd)
			_ = cmd.Run()
		case "darwin":
			cmd := exec.CommandContext(commandCtx, "open", href)
			Setpgid(cmd)
			_ = cmd.Run()
		}
	}
}

// actionShellWithDir gets an ActionFunc that runs a shell command in a directory.
func actionShellWithDir(pluginDir string, terminal bool, appleScriptTemplate, command string, params, envVars []string) ActionFunc {
	if terminal {
		return actionShellTerminalDirect(pluginDir, appleScriptTemplate, command, params, envVars)
	}
	return func(ctx context.Context) {
		cmd := exec.CommandContext(ctx, command, params...)
		Setpgid(cmd)
		cmd.Dir = pluginDir
		cmd.Env = append(cmd.Env, os.Environ()...)
		_ = cmd.Run()
	}
}

// actionShellTerminalDirect runs shell commands where terminal=true.
func actionShellTerminalDirect(pluginDir, appleScriptTemplate, command string, params, envVars []string) ActionFunc {
	return func(ctx context.Context) {
		quotedCmd := strconv.Quote(command)
		quotedCmd = quotedCmd[1 : len(quotedCmd)-1]
		quotedParams := make([]string, len(params))
		for i := range params {
			quotedParams[i] = strconv.Quote(params[i])
			quotedParams[i] = quotedParams[i][1 : len(quotedParams[i])-1]
		}
		paramsStr := strconv.Quote(strings.Join(quotedParams, " "))
		_ = runInTerminalDirect(appleScriptTemplate, quotedCmd, paramsStr, envVars)
	}
}

// actionRefresh gets an ActionFunc that manually refreshes the
// Plugin.
func actionRefresh(refreshFunc func(ctx context.Context)) ActionFunc {
	return func(ctx context.Context) {
		refreshFunc(ctx)
	}
}
