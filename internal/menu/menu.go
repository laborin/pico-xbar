package menu

import (
	"context"
	"encoding/base64"
	"os"
	"os/exec"
	"sync"

	"github.com/laborin/pico-xbar/internal/statusbar"
	"github.com/laborin/pico-xbar/pkg/plugins"
)

const (
	menuIndexRefresh    = -1
	menuIndexOpenPlugin = -2
	menuIndexQuit       = -3
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
	itemId := statusbar.CreateStatusItem()

	pm := &PluginMenu{
		itemId:    itemId,
		plugin:    p,
		ctx:       context.Background(),
		menuItems: make(map[int]*plugins.Item),
	}

	return pm
}

func (pm *PluginMenu) Setup() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.updateTitle()
	pm.rebuildMenu()
}

func (pm *PluginMenu) Update(ctx context.Context, p *plugins.Plugin) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.ctx = ctx
	pm.plugin = p

	pm.updateTitle()
	pm.rebuildMenu()
}

func (pm *PluginMenu) UpdateLabel(p *plugins.Plugin) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

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

	for _, item := range pm.plugin.Items.ExpandedItems {
		pm.addMenuItem(item, "")
	}

	statusbar.AddMenuItem(pm.itemId, -100, "", false, true)
	statusbar.AddMenuItem(pm.itemId, menuIndexRefresh, "Refresh", false, false)
	statusbar.AddMenuItem(pm.itemId, menuIndexOpenPlugin, "Open Plugin...", false, false)
	statusbar.AddMenuItem(pm.itemId, -101, "", false, true)
	statusbar.AddMenuItem(pm.itemId, menuIndexQuit, "Quit", false, false)
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
		if parentSubmenu == "" {
			statusbar.AddMenuItemWithColor(pm.itemId, index, text, item.Params.Disabled, false, item.Params.Color)
		} else {
			statusbar.AddSubmenuItem(pm.itemId, parentSubmenu, index, text, item.Params.Disabled, false, item.Params.Color)
		}

		if item.Params.TemplateImage != "" {
			if iconBytes, err := base64.StdEncoding.DecodeString(item.Params.TemplateImage); err == nil {
				statusbar.SetMenuItemIcon(pm.itemId, index, iconBytes, true)
			}
		} else if item.Params.Image != "" {
			if iconBytes, err := base64.StdEncoding.DecodeString(item.Params.Image); err == nil {
				statusbar.SetMenuItemIcon(pm.itemId, index, iconBytes, false)
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
