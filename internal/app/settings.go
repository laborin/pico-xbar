package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
)

type Settings struct {
	sync.Mutex

	path string `json:"-"`

	Terminal struct {
		AppleScriptTemplate3 string `json:"appleScriptTemplate3"`
	} `json:"terminal"`
}

func (s *Settings) setDefaults() {
	if s.Terminal.AppleScriptTemplate3 == "" {
		s.Terminal.AppleScriptTemplate3 = `
			set quotedScriptName to quoted form of "{{ .Command }}"
		{{ if .Params }}
			set commandLine to {{ .Vars }} & " " & quotedScriptName & " " & {{ .Params }}
		{{ else }}
			set commandLine to {{ .Vars }} & " " & quotedScriptName
		{{ end }}
			if application "Terminal" is running then
				tell application "Terminal"
					do script commandLine
					activate
				end tell
			else
				tell application "Terminal"
					do script commandLine in window 1
					activate
				end tell
			end if
		`
	}
}

func LoadSettings(path string) (*Settings, error) {
	s := &Settings{
		path: path,
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.setDefaults()
			return s, nil
		}
		return nil, errors.Wrap(err, "ReadFile")
	}
	err = json.Unmarshal(b, s)
	if err != nil {
		return nil, errors.Wrap(err, "Unmarshal")
	}
	s.setDefaults()
	return s, nil
}

func (s *Settings) Save() error {
	s.Lock()
	defer s.Unlock()
	s.setDefaults()
	b, err := json.MarshalIndent(s, "", "\t")
	if err != nil {
		return errors.Wrap(err, "MarshalIndent")
	}
	err = os.MkdirAll(filepath.Dir(s.path), 0777)
	if err != nil {
		return errors.Wrap(err, "MkdirAll")
	}
	err = os.WriteFile(s.path, b, 0777)
	if err != nil {
		return errors.Wrap(err, "WriteFile")
	}
	return nil
}

func (s *Settings) AppleScriptTemplate() string {
	return s.Terminal.AppleScriptTemplate3
}
