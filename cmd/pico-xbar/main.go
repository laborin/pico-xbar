package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/laborin/pico-xbar/internal/app"
	"github.com/laborin/pico-xbar/internal/menu"
)

func main() {
	log.SetFlags(log.LstdFlags)

	dataDir, err := getDataDir()
	if err != nil {
		log.Fatal("failed to get data directory:", err)
	}

	application, err := app.New(dataDir)
	if err != nil {
		log.Fatal("failed to create app:", err)
	}

	menu.Run(application.Start, application.Stop)
}

func getDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "xbar"), nil
}
