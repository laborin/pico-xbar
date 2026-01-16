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

type pluginEntry struct {
	plugin *plugins.Plugin
	menu   *menu.PluginMenu
	cancel context.CancelFunc
}

type App struct {
	dataDir   string
	pluginDir string
	settings  *Settings

	plugins       map[string]*pluginEntry
	pluginsLock   sync.Mutex
	defaultMenuId int

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
	a.plugins = make(map[string]*pluginEntry)
	a.defaultMenuId = -1

	statusbar.SetClickHandler(a.handleClick)
	menu.SetRefreshAllHandler(a.RefreshAll)

	pluginList := a.loadPlugins()
	if len(pluginList) == 0 {
		a.showDefaultMenu()
		return
	}

	for _, p := range pluginList {
		a.startPlugin(p)
	}
}

func (a *App) showDefaultMenu() {
	if a.defaultMenuId == -1 {
		a.defaultMenuId = menu.ShowDefault(a.pluginDir, a.RefreshAll, a.Stop)
	}
}

func (a *App) hideDefaultMenu() {
	if a.defaultMenuId != -1 {
		statusbar.RemoveStatusItem(a.defaultMenuId)
		a.defaultMenuId = -1
	}
}

func (a *App) Stop() {
	if a.cancel != nil {
		a.cancel()
	}
	a.wg.Wait()
}

func (a *App) loadPlugins() plugins.Plugins {
	pluginList, err := plugins.Dir(a.pluginDir)
	if err != nil {
		log.Printf("failed to load plugins: %v", err)
		return nil
	}
	log.Printf("loaded %d plugins", len(pluginList))
	return pluginList
}

func (a *App) startPlugin(p *plugins.Plugin) {
	p.AppleScriptTemplate = a.settings.AppleScriptTemplate()
	p.Debugf = plugins.DebugfLog

	pm := menu.NewPluginMenu(p)
	pluginCtx, pluginCancel := context.WithCancel(a.ctx)

	entry := &pluginEntry{
		plugin: p,
		menu:   pm,
		cancel: pluginCancel,
	}

	a.pluginsLock.Lock()
	a.plugins[p.Command] = entry
	a.pluginsLock.Unlock()

	p.OnRefresh = func(ctx context.Context, plugin *plugins.Plugin, err error) {
		pm.Update(ctx, plugin)
	}

	p.OnCycle = func(ctx context.Context, plugin *plugins.Plugin) {
		pm.UpdateLabel(plugin)
	}

	p.OnRemove = func() {
		a.removePlugin(p.Command)
	}

	pm.Setup()

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		p.Run(pluginCtx)
	}()
}

func (a *App) removePlugin(command string) {
	a.pluginsLock.Lock()
	entry, exists := a.plugins[command]
	if exists {
		delete(a.plugins, command)
	}
	isEmpty := len(a.plugins) == 0
	a.pluginsLock.Unlock()

	if exists && entry != nil {
		entry.cancel()
		if entry.plugin != nil {
			entry.plugin.OnRefresh = nil
			entry.plugin.OnCycle = nil
			entry.plugin.OnRemove = nil
		}
		if entry.menu != nil {
			entry.menu.Remove()
		}
		log.Printf("removed plugin: %s", filepath.Base(command))
	}

	if isEmpty {
		a.showDefaultMenu()
	}
}

func (a *App) handleClick(itemId int, menuItemIndex int) {
	a.pluginsLock.Lock()
	var pm *menu.PluginMenu
	for _, entry := range a.plugins {
		if entry.menu != nil && entry.menu.ItemId() == itemId {
			pm = entry.menu
			break
		}
	}
	a.pluginsLock.Unlock()

	if pm != nil {
		pm.HandleClick(menuItemIndex)
	}
}

func (a *App) RefreshAll() {
	log.Printf("refreshing all plugins...")

	newPluginList := a.loadPlugins()
	newPluginMap := make(map[string]*plugins.Plugin)
	for _, p := range newPluginList {
		newPluginMap[p.Command] = p
	}

	a.pluginsLock.Lock()
	var toRemove []string
	for cmd := range a.plugins {
		if _, exists := newPluginMap[cmd]; !exists {
			toRemove = append(toRemove, cmd)
		}
	}
	a.pluginsLock.Unlock()

	for _, cmd := range toRemove {
		a.removePlugin(cmd)
	}

	if len(newPluginMap) > 0 {
		a.hideDefaultMenu()
	}

	for cmd, p := range newPluginMap {
		a.pluginsLock.Lock()
		_, exists := a.plugins[cmd]
		a.pluginsLock.Unlock()

		if !exists {
			a.startPlugin(p)
		}
	}

	if len(newPluginMap) == 0 {
		a.showDefaultMenu()
	}

	log.Printf("refresh all complete: %d plugins", len(newPluginMap))
}
