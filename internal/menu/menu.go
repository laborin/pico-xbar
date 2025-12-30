package menu

import (
	"context"
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/laborin/pico-xbar/internal/loginitem"
	"github.com/laborin/pico-xbar/internal/statusbar"
	"github.com/laborin/pico-xbar/pkg/plugins"
)

const (
	menuIndexRefresh        = -1
	menuIndexOpenPlugin     = -2
	menuIndexQuit           = -3
	menuIndexCopyPath       = -4
	menuIndexShowInFinder   = -5
	menuIndexOpenTerminal   = -6
	menuIndexStartAtLogin   = -7
)

type PluginMenu struct {
	itemId    int
	plugin    *plugins.Plugin
	ctx       context.Context
	mu        sync.Mutex
	menuItems map[int]*plugins.Item
	nextIndex int
}

func NewPluginMenu(p *plugins.Plugin) *PluginMenu {
	pm := &PluginMenu{
		itemId:    -1,
		plugin:    p,
		ctx:       context.Background(),
		menuItems: make(map[int]*plugins.Item),
	}

	return pm
}

func (pm *PluginMenu) Setup() {
}

func (pm *PluginMenu) Update(ctx context.Context, p *plugins.Plugin) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.ctx = ctx
	pm.plugin = p

	if pm.itemId == -1 {
		pm.itemId = statusbar.CreateStatusItem()
	}

	pm.updateTitle()
	pm.rebuildMenu()
}

func (pm *PluginMenu) UpdateLabel(p *plugins.Plugin) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.itemId == -1 {
		return
	}

	pm.plugin = p
	pm.updateTitle()
}

func (pm *PluginMenu) updateTitle() {
	item := pm.plugin.CurrentCycleItem()
	if item == nil {
		statusbar.SetTitle(pm.itemId, pm.plugin.CleanFilename())
		return
	}

	statusbar.SetTitle(pm.itemId, item.DisplayText())

	if item.Params.TemplateImage != "" {
		if iconBytes, err := base64.StdEncoding.DecodeString(item.Params.TemplateImage); err == nil {
			statusbar.SetIcon(pm.itemId, iconBytes, true)
		}
	} else if item.Params.Image != "" {
		if iconBytes, err := base64.StdEncoding.DecodeString(item.Params.Image); err == nil {
			statusbar.SetIcon(pm.itemId, iconBytes, false)
		}
	}
}

func (pm *PluginMenu) rebuildMenu() {
	statusbar.ClearMenu(pm.itemId)
	pm.menuItems = make(map[int]*plugins.Item)
	pm.nextIndex = 0

	items := pm.plugin.Items.ExpandedItems
	for _, item := range items {
		pm.addMenuItem(item, "")
	}

	pluginSubmenu := "Plugin..."

	if len(items) == 0 {
		statusbar.AddMenuItem(pm.itemId, menuIndexRefresh, "Refresh", false, false)
		statusbar.AddSubmenu(pm.itemId, pluginSubmenu)
		pm.addPluginSubmenuItems(pluginSubmenu)
		statusbar.AddMenuItem(pm.itemId, menuIndexStartAtLogin, "Start at Login", false, false)
		statusbar.SetMenuItemState(pm.itemId, menuIndexStartAtLogin, loginitem.IsEnabled())
		statusbar.AddMenuItem(pm.itemId, -101, "", false, true)
		statusbar.AddMenuItem(pm.itemId, menuIndexQuit, "Quit", false, false)
	} else {
		lastIsSeparator := len(items) > 0 && items[len(items)-1].Params.Separator
		if !lastIsSeparator {
			statusbar.AddMenuItem(pm.itemId, -100, "", false, true)
		}
		statusbar.AddSubmenu(pm.itemId, "pico-xbar")
		statusbar.AddSubmenuItem(pm.itemId, "pico-xbar", menuIndexRefresh, "Refresh", false, false, "")
		statusbar.AddNestedSubmenu(pm.itemId, "pico-xbar", pluginSubmenu)
		pm.addPluginSubmenuItems(pluginSubmenu)
		statusbar.AddSubmenuItem(pm.itemId, "pico-xbar", menuIndexStartAtLogin, "Start at Login", false, false, "")
		statusbar.SetMenuItemState(pm.itemId, menuIndexStartAtLogin, loginitem.IsEnabled())
		statusbar.AddSubmenuItem(pm.itemId, "pico-xbar", -101, "", false, true, "")
		statusbar.AddSubmenuItem(pm.itemId, "pico-xbar", menuIndexQuit, "Quit", false, false, "")
	}
}

func (pm *PluginMenu) addPluginSubmenuItems(submenuName string) {
	pluginName := pm.plugin.CleanFilename()
	statusbar.AddSubmenuItem(pm.itemId, submenuName, -200, pluginName, true, false, "")
	statusbar.AddSubmenuItem(pm.itemId, submenuName, -201, "", false, true, "")
	statusbar.AddSubmenuItem(pm.itemId, submenuName, menuIndexOpenPlugin, "Open Plugin File...", false, false, "")
	statusbar.AddSubmenuItem(pm.itemId, submenuName, menuIndexCopyPath, "Copy Plugin Path", false, false, "")
	statusbar.AddSubmenuItem(pm.itemId, submenuName, menuIndexShowInFinder, "Show in Finder", false, false, "")
	statusbar.AddSubmenuItem(pm.itemId, submenuName, menuIndexOpenTerminal, "Open Terminal Here...", false, false, "")
}

func (pm *PluginMenu) addMenuItem(item *plugins.Item, parentSubmenu string) {
	index := pm.nextIndex
	pm.nextIndex++

	if item.Params.Separator {
		if parentSubmenu == "" {
			statusbar.AddMenuItem(pm.itemId, index, "", false, true)
		} else {
			statusbar.AddSubmenuItem(pm.itemId, parentSubmenu, index, "", false, true, "")
		}
		return
	}

	text := item.DisplayText()
	hasChildren := len(item.Items) > 0

	if hasChildren {
		if parentSubmenu == "" {
			statusbar.AddSubmenu(pm.itemId, text)
		} else {
			statusbar.AddNestedSubmenu(pm.itemId, parentSubmenu, text)
		}

		for _, child := range item.Items {
			pm.addMenuItem(child, text)
		}
	} else {
		disabled := item.Params.Disabled || item.Action() == nil
		if parentSubmenu == "" {
			statusbar.AddMenuItemWithColor(pm.itemId, index, text, disabled, false, item.Params.Color)
		} else {
			statusbar.AddSubmenuItem(pm.itemId, parentSubmenu, index, text, disabled, false, item.Params.Color)
		}

		if item.Params.TemplateImage != "" {
			if iconBytes, err := base64.StdEncoding.DecodeString(item.Params.TemplateImage); err == nil {
				statusbar.SetMenuItemIcon(pm.itemId, index, iconBytes, true, item.Params.ShrinkImage)
			}
		} else if item.Params.Image != "" {
			if iconBytes, err := base64.StdEncoding.DecodeString(item.Params.Image); err == nil {
				statusbar.SetMenuItemIcon(pm.itemId, index, iconBytes, false, item.Params.ShrinkImage)
			}
		}

		pm.menuItems[index] = item
	}
}

func (pm *PluginMenu) HandleClick(menuItemIndex int) {
	pm.mu.Lock()
	ctx := pm.ctx
	plugin := pm.plugin
	item := pm.menuItems[menuItemIndex]
	pm.mu.Unlock()

	switch menuItemIndex {
	case menuIndexRefresh:
		plugin.TriggerRefresh()
	case menuIndexOpenPlugin:
		exec.Command("open", plugin.Command).Start()
	case menuIndexCopyPath:
		statusbar.CopyToClipboard(plugin.Command)
		statusbar.ShowAlert("Path Copied", "Plugin path copied to clipboard.")
	case menuIndexShowInFinder:
		exec.Command("open", "-R", plugin.Command).Start()
	case menuIndexOpenTerminal:
		pluginDir := filepath.Dir(plugin.Command)
		exec.Command("open", "-a", "Terminal", pluginDir).Start()
	case menuIndexStartAtLogin:
		if err := loginitem.Toggle(); err != nil {
			statusbar.ShowAlert("Error", "Failed to toggle login item: "+err.Error())
		}
		statusbar.SetMenuItemState(pm.itemId, menuIndexStartAtLogin, loginitem.IsEnabled())
	case menuIndexQuit:
		statusbar.Stop()
	default:
		if item != nil {
			action := item.Action()
			if action != nil {
				action(ctx)
			}
		}
	}
}

func (pm *PluginMenu) ItemId() int {
	return pm.itemId
}

func ShowDefault(pluginDir string, onQuit func()) int {
	itemId := statusbar.CreateStatusItem()
	statusbar.SetTitle(itemId, "xbar")

	statusbar.AddMenuItem(itemId, 0, "Open Plugin Folder...", false, false)
	statusbar.AddMenuItem(itemId, -1, "", false, true)
	statusbar.AddMenuItem(itemId, 1, "Quit", false, false)

	statusbar.SetClickHandler(func(id int, menuIndex int) {
		if id != itemId {
			return
		}
		switch menuIndex {
		case 0:
			os.MkdirAll(pluginDir, 0755)
			exec.Command("open", pluginDir).Start()
		case 1:
			if onQuit != nil {
				onQuit()
			}
			statusbar.Stop()
		}
	})

	return itemId
}

func Run(onReady func(), onExit func()) {
	statusbar.Run(onReady)
	if onExit != nil {
		onExit()
	}
}
