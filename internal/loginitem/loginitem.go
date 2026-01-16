package loginitem

import (
	"os"
	"path/filepath"
)

const (
	bundleID     = "com.laborin.pico-xbar"
	plistName    = bundleID + ".plist"
	plistContent = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>` + bundleID + `</string>
	<key>ProgramArguments</key>
	<array>
		<string>/usr/bin/open</string>
		<string>-a</string>
		<string>pico-xbar</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>
`
)

func launchAgentPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", plistName), nil
}

func IsEnabled() bool {
	path, err := launchAgentPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func Enable() error {
	path, err := launchAgentPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(plistContent), 0644)
}

func Disable() error {
	path, err := launchAgentPath()
	if err != nil {
		return err
	}

	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func Toggle() error {
	if IsEnabled() {
		return Disable()
	}
	return Enable()
}
