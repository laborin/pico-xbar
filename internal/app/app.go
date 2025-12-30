package app

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/laborin/pico-xbar/internal/menu"
	"github.com/laborin/pico-xbar/internal/statusbar"
	"github.com/laborin/pico-xbar/pkg/plugins"
)

type App struct {
	dataDir   string
	pluginDir string
	settings  *Settings

	plugins     plugins.Plugins
	pluginMenus []*menu.PluginMenu
	menuLock    sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func New(dataDir string) (*App, error) {
	pluginDir := filepath.Join(dataDir, "plugins")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return nil, err
	}

	settingsPath := filepath.Join(dataDir, "xbar.config.json")
	settings, err := LoadSettings(settingsPath)
	if err != nil {
		log.Printf("warning: failed to load settings: %v", err)
		settings = &Settings{}
	}

	return &App{
		dataDir:   dataDir,
		pluginDir: pluginDir,
		settings:  settings,
	}, nil
}

func (a *App) Start() {
	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.loadPlugins()

	statusbar.SetClickHandler(a.handleClick)

	if len(a.plugins) == 0 {
		menu.ShowDefault(a.pluginDir, a.Stop)
		return
	}

	for _, p := range a.plugins {
		a.startPlugin(p)
	}
}

func (a *App) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
	a.wg.Wait()
}

func (a *App) loadPlugins() {
	var err error
	a.plugins, err = plugins.Dir(a.pluginDir)
	if err != nil {
		log.Printf("failed to load plugins: %v", err)
		return
	}
	log.Printf("loaded %d plugins", len(a.plugins))
}

func (a *App) startPlugin(p *plugins.Plugin) {
	p.AppleScriptTemplate = a.settings.AppleScriptTemplate()
	p.Debugf = plugins.DebugfLog

	pm := menu.NewPluginMenu(p)

	a.menuLock.Lock()
	a.pluginMenus = append(a.pluginMenus, pm)
	a.menuLock.Unlock()

	p.OnRefresh = func(ctx context.Context, plugin *plugins.Plugin, err error) {
		pm.Update(ctx, plugin)
	}

	p.OnCycle = func(ctx context.Context, plugin *plugins.Plugin) {
		pm.UpdateLabel(plugin)
	}

	pm.Setup()

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		p.Run(a.ctx)
	}()
}

func (a *App) handleClick(itemId int, menuItemIndex int) {
	a.menuLock.Lock()
	var pm *menu.PluginMenu
	for _, m := range a.pluginMenus {
		if m.ItemId() == itemId {
			pm = m
			break
		}
	}
	a.menuLock.Unlock()

	if pm != nil {
		pm.HandleClick(menuItemIndex)
	}
}
