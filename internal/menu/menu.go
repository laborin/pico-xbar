package menu

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/laborin/pico-xbar/internal/loginitem"
	"github.com/laborin/pico-xbar/internal/statusbar"
	"github.com/laborin/pico-xbar/internal/version"
	"github.com/laborin/pico-xbar/pkg/plugins"
)

func L(key string) string {
	return statusbar.LocalizedString(key)
}

const (
	menuIndexRefresh      = -1
	menuIndexOpenPlugin   = -2
	menuIndexQuit         = -3
	menuIndexCopyPath     = -4
	menuIndexShowInFinder = -5
	menuIndexOpenTerminal = -6
	menuIndexStartAtLogin = -7
	menuIndexAbout        = -8
	menuIndexRefreshAll   = -9
)

var globalRefreshAllFunc func()

func SetRefreshAllHandler(f func()) {
	globalRefreshAllFunc = f
}

func ClearRefreshAllHandler() {
	globalRefreshAllFunc = nil
}

type menuItemState struct {
	tag           int
	text          string
	disabled      bool
	separator     bool
	isSubmenu     bool
	submenuTitle  string
	color         string
	font          string
	size          int
	key           string
	alternate     bool
	imageHash     string
	templateImage bool
	shrinkImage   bool
	children      []*menuItemState
}

func (s *menuItemState) equals(other *menuItemState) bool {
	if s == nil || other == nil {
		return s == other
	}
	return s.text == other.text &&
		s.disabled == other.disabled &&
		s.separator == other.separator &&
		s.isSubmenu == other.isSubmenu &&
		s.color == other.color &&
		s.font == other.font &&
		s.size == other.size &&
		s.key == other.key &&
		s.alternate == other.alternate &&
		s.imageHash == other.imageHash
}

type PluginMenu struct {
	itemId        int
	plugin        *plugins.Plugin
	ctx           context.Context
	mu            sync.Mutex
	menuItems     map[int]*plugins.Item
	nextIndex     int
	prevState     []*menuItemState
	hadItems      bool
	prevCommand   string
	prevTitle     string
	prevImageHash string
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

	commandChanged := pm.prevCommand != "" && pm.prevCommand != p.Command
	pm.prevCommand = p.Command

	hasItems := len(pm.plugin.Items.ExpandedItems) > 0
	if pm.hadItems != hasItems || commandChanged {
		pm.rebuildMenu()
		return
	}

	newState := pm.buildState(pm.plugin.Items.ExpandedItems)
	if pm.prevState == nil {
		pm.rebuildMenu()
		return
	}

	pm.menuItems = make(map[int]*plugins.Item)
	pm.diffAndUpdate(pm.prevState, newState, "")
	pm.prevState = newState
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
		title := pm.plugin.CleanFilename()
		if title != pm.prevTitle {
			statusbar.SetTitle(pm.itemId, title)
			pm.prevTitle = title
		}
		return
	}

	title := item.DisplayText()
	if title != pm.prevTitle {
		statusbar.SetTitle(pm.itemId, title)
		pm.prevTitle = title
	}

	var imageHash string
	if item.Params.TemplateImage != "" {
		imageHash = "t:" + item.Params.TemplateImage[:min(32, len(item.Params.TemplateImage))]
		if imageHash != pm.prevImageHash {
			if iconBytes, err := base64.StdEncoding.DecodeString(item.Params.TemplateImage); err == nil {
				statusbar.SetIcon(pm.itemId, iconBytes, true)
			}
			pm.prevImageHash = imageHash
		}
	} else if item.Params.Image != "" {
		imageHash = "i:" + item.Params.Image[:min(32, len(item.Params.Image))]
		if imageHash != pm.prevImageHash {
			if iconBytes, err := base64.StdEncoding.DecodeString(item.Params.Image); err == nil {
				statusbar.SetIcon(pm.itemId, iconBytes, false)
			}
			pm.prevImageHash = imageHash
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

	pluginSubmenu := L("Plugin...")

	if len(items) == 0 {
		statusbar.AddMenuItem(pm.itemId, menuIndexRefresh, L("Refresh"), false, false)
		statusbar.AddMenuItem(pm.itemId, menuIndexRefreshAll, L("Refresh All"), false, false)
		statusbar.AddSubmenu(pm.itemId, pluginSubmenu)
		pm.addPluginSubmenuItems(pluginSubmenu)
		statusbar.AddMenuItem(pm.itemId, menuIndexStartAtLogin, L("Start at Login"), false, false)
		statusbar.SetMenuItemState(pm.itemId, menuIndexStartAtLogin, loginitem.IsEnabled())
		statusbar.AddMenuItem(pm.itemId, menuIndexAbout, L("About pico-xbar"), false, false)
		statusbar.AddMenuItem(pm.itemId, -101, "", false, true)
		statusbar.AddMenuItem(pm.itemId, menuIndexQuit, L("Quit"), false, false)
	} else {
		lastIsSeparator := len(items) > 0 && items[len(items)-1].Params.Separator
		if !lastIsSeparator {
			statusbar.AddMenuItem(pm.itemId, -100, "", false, true)
		}
		statusbar.AddSubmenu(pm.itemId, "pico-xbar")
		statusbar.AddSubmenuItem(pm.itemId, "pico-xbar", menuIndexRefresh, L("Refresh"), false, false, "")
		statusbar.AddSubmenuItem(pm.itemId, "pico-xbar", menuIndexRefreshAll, L("Refresh All"), false, false, "")
		statusbar.AddNestedSubmenu(pm.itemId, "pico-xbar", pluginSubmenu)
		pm.addPluginSubmenuItems(pluginSubmenu)
		statusbar.AddSubmenuItem(pm.itemId, "pico-xbar", menuIndexStartAtLogin, L("Start at Login"), false, false, "")
		statusbar.SetMenuItemState(pm.itemId, menuIndexStartAtLogin, loginitem.IsEnabled())
		statusbar.AddSubmenuItem(pm.itemId, "pico-xbar", menuIndexAbout, L("About pico-xbar"), false, false, "")
		statusbar.AddSubmenuItem(pm.itemId, "pico-xbar", -101, "", false, true, "")
		statusbar.AddSubmenuItem(pm.itemId, "pico-xbar", menuIndexQuit, L("Quit"), false, false, "")
	}

	pm.prevState = pm.buildState(items)
	pm.hadItems = len(items) > 0
}

func (pm *PluginMenu) addPluginSubmenuItems(submenuName string) {
	pluginName := pm.plugin.CleanFilename()
	statusbar.AddSubmenuItem(pm.itemId, submenuName, -200, pluginName, true, false, "")
	statusbar.AddSubmenuItem(pm.itemId, submenuName, -201, "", false, true, "")
	statusbar.AddSubmenuItem(pm.itemId, submenuName, menuIndexOpenPlugin, L("Open Plugin File..."), false, false, "")
	statusbar.AddSubmenuItem(pm.itemId, submenuName, menuIndexCopyPath, L("Copy Plugin Path"), false, false, "")
	statusbar.AddSubmenuItem(pm.itemId, submenuName, menuIndexShowInFinder, L("Show in Finder"), false, false, "")
	statusbar.AddSubmenuItem(pm.itemId, submenuName, menuIndexOpenTerminal, L("Open Terminal Here..."), false, false, "")
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
		p := item.Params
		if parentSubmenu == "" {
			statusbar.AddMenuItemStyled(pm.itemId, index, text, disabled, false, p.Color, p.Font, p.Size, p.Key, p.Alternate)
		} else {
			statusbar.AddSubmenuItemStyled(pm.itemId, parentSubmenu, index, text, disabled, false, p.Color, p.Font, p.Size, p.Key, p.Alternate)
		}

		if p.TemplateImage != "" {
			if iconBytes, err := base64.StdEncoding.DecodeString(p.TemplateImage); err == nil {
				statusbar.SetMenuItemIcon(pm.itemId, index, iconBytes, true, p.ShrinkImage)
			}
		} else if p.Image != "" {
			if iconBytes, err := base64.StdEncoding.DecodeString(p.Image); err == nil {
				statusbar.SetMenuItemIcon(pm.itemId, index, iconBytes, false, p.ShrinkImage)
			}
		}

		pm.menuItems[index] = item
	}
}

func (pm *PluginMenu) buildState(items []*plugins.Item) []*menuItemState {
	states := make([]*menuItemState, 0, len(items))
	for _, item := range items {
		state := pm.buildItemState(item)
		states = append(states, state)
	}
	return states
}

func (pm *PluginMenu) buildItemState(item *plugins.Item) *menuItemState {
	state := &menuItemState{
		text:      item.DisplayText(),
		separator: item.Params.Separator,
		disabled:  item.Params.Disabled || item.Action() == nil,
		color:     item.Params.Color,
		font:      item.Params.Font,
		size:      item.Params.Size,
		key:       item.Params.Key,
		alternate: item.Params.Alternate,
	}

	if item.Params.TemplateImage != "" {
		state.imageHash = fmt.Sprintf("%x", md5.Sum([]byte(item.Params.TemplateImage)))
		state.templateImage = true
	} else if item.Params.Image != "" {
		state.imageHash = fmt.Sprintf("%x", md5.Sum([]byte(item.Params.Image)))
		state.templateImage = false
	}
	state.shrinkImage = item.Params.ShrinkImage

	if len(item.Items) > 0 {
		state.isSubmenu = true
		state.submenuTitle = state.text
		for _, child := range item.Items {
			state.children = append(state.children, pm.buildItemState(child))
		}
	}

	return state
}

func (pm *PluginMenu) diffAndUpdate(oldState, newState []*menuItemState, parentSubmenu string) {
	oldLen := len(oldState)
	newLen := len(newState)
	minLen := oldLen
	if newLen < minLen {
		minLen = newLen
	}

	for i := oldLen - 1; i >= newLen; i-- {
		if parentSubmenu == "" {
			statusbar.RemoveMenuItemAtIndex(pm.itemId, i)
		} else {
			statusbar.RemoveSubmenuItemAtIndex(pm.itemId, parentSubmenu, i)
		}
	}

	for i := 0; i < minLen; i++ {
		oldItem := oldState[i]
		newItem := newState[i]

		if oldItem.isSubmenu && newItem.isSubmenu && oldItem.submenuTitle == newItem.submenuTitle {
			pm.diffAndUpdate(oldItem.children, newItem.children, newItem.submenuTitle)
			continue
		}

		needsReplace := oldItem.isSubmenu != newItem.isSubmenu ||
			oldItem.separator != newItem.separator ||
			(oldItem.isSubmenu && newItem.isSubmenu && oldItem.submenuTitle != newItem.submenuTitle)

		if needsReplace {
			if parentSubmenu == "" {
				statusbar.RemoveMenuItemAtIndex(pm.itemId, i)
			} else {
				statusbar.RemoveSubmenuItemAtIndex(pm.itemId, parentSubmenu, i)
			}
			pm.insertItemAtIndex(newItem, i, parentSubmenu)
			continue
		}

		if !oldItem.equals(newItem) {
			pm.updateItemAtIndex(newItem, i, parentSubmenu)
		}
	}

	for i := oldLen; i < newLen; i++ {
		pm.insertItemAtIndex(newState[i], i, parentSubmenu)
	}
}

func (pm *PluginMenu) insertItemAtIndex(state *menuItemState, atIndex int, parentSubmenu string) {
	tag := pm.nextIndex
	pm.nextIndex++

	if state.separator {
		if parentSubmenu == "" {
			statusbar.InsertMenuItemStyledAtIndex(pm.itemId, atIndex, tag, "", false, true, "", "", 0, "", false)
		} else {
			statusbar.InsertSubmenuItemStyledAtIndex(pm.itemId, parentSubmenu, atIndex, tag, "", false, true, "", "", 0, "", false)
		}
		return
	}

	if state.isSubmenu {
		if parentSubmenu == "" {
			statusbar.InsertSubmenuAtIndex(pm.itemId, atIndex, state.submenuTitle)
		} else {
			statusbar.InsertNestedSubmenuAtIndex(pm.itemId, parentSubmenu, atIndex, state.submenuTitle)
		}
		for j, child := range state.children {
			pm.insertItemAtIndex(child, j, state.submenuTitle)
		}
		return
	}

	if parentSubmenu == "" {
		statusbar.InsertMenuItemStyledAtIndex(pm.itemId, atIndex, tag, state.text, state.disabled, false, state.color, state.font, state.size, state.key, state.alternate)
	} else {
		statusbar.InsertSubmenuItemStyledAtIndex(pm.itemId, parentSubmenu, atIndex, tag, state.text, state.disabled, false, state.color, state.font, state.size, state.key, state.alternate)
	}

	if state.imageHash != "" {
		pm.setImageForState(state, tag)
	}

	state.tag = tag
}

func (pm *PluginMenu) updateItemAtIndex(state *menuItemState, atIndex int, parentSubmenu string) {
	tag := pm.nextIndex
	pm.nextIndex++

	if state.separator {
		return
	}

	if parentSubmenu == "" {
		statusbar.UpdateMenuItemAtIndex(pm.itemId, atIndex, tag, state.text, state.disabled, state.separator, state.color, state.font, state.size, state.key, state.alternate)
	} else {
		statusbar.UpdateSubmenuItemAtIndex(pm.itemId, parentSubmenu, atIndex, tag, state.text, state.disabled, state.separator, state.color, state.font, state.size, state.key, state.alternate)
	}

	if state.imageHash != "" {
		pm.setImageForState(state, tag)
	}

	state.tag = tag
}

func (pm *PluginMenu) setImageForState(state *menuItemState, tag int) {
	items := pm.plugin.Items.ExpandedItems
	item := pm.findItemByText(items, state.text)
	if item == nil {
		return
	}

	if item.Params.TemplateImage != "" {
		if iconBytes, err := base64.StdEncoding.DecodeString(item.Params.TemplateImage); err == nil {
			statusbar.SetMenuItemIcon(pm.itemId, tag, iconBytes, true, state.shrinkImage)
		}
	} else if item.Params.Image != "" {
		if iconBytes, err := base64.StdEncoding.DecodeString(item.Params.Image); err == nil {
			statusbar.SetMenuItemIcon(pm.itemId, tag, iconBytes, false, state.shrinkImage)
		}
	}

	pm.menuItems[tag] = item
}

func (pm *PluginMenu) findItemByText(items []*plugins.Item, text string) *plugins.Item {
	for _, item := range items {
		if item.DisplayText() == text {
			return item
		}
		if len(item.Items) > 0 {
			if found := pm.findItemByText(item.Items, text); found != nil {
				return found
			}
		}
	}
	return nil
}

func (pm *PluginMenu) HandleClick(menuItemIndex int) {
	pm.mu.Lock()
	ctx := pm.ctx
	plugin := pm.plugin
	item := pm.menuItems[menuItemIndex]
	removed := pm.itemId == -1
	pm.mu.Unlock()

	switch menuItemIndex {
	case menuIndexRefreshAll:
		if globalRefreshAllFunc != nil {
			go globalRefreshAllFunc()
		}
	case menuIndexStartAtLogin:
		if err := loginitem.Toggle(); err != nil {
			statusbar.ShowAlert(L("Error"), "Failed to toggle login item: "+err.Error())
		}
		if !removed {
			statusbar.SetMenuItemState(pm.itemId, menuIndexStartAtLogin, loginitem.IsEnabled())
		}
	case menuIndexAbout:
		aboutMsg := strings.Replace(L("about_message"), "%@", version.Version, 1)
		statusbar.ShowAlertWithURL("pico-xbar", aboutMsg, L("Website"), "https://laborin.com.mx/pico-xbar")
	case menuIndexQuit:
		statusbar.Stop()
	case menuIndexRefresh:
		if !removed {
			plugin.TriggerRefresh()
		}
	case menuIndexOpenPlugin:
		if !removed {
			runCommandDetached("open", plugin.Command)
		}
	case menuIndexCopyPath:
		if !removed {
			statusbar.CopyToClipboard(plugin.Command)
			statusbar.ShowAlert(L("Path Copied"), L("Plugin path copied to clipboard."))
		}
	case menuIndexShowInFinder:
		if !removed {
			runCommandDetached("open", "-R", plugin.Command)
		}
	case menuIndexOpenTerminal:
		if !removed {
			pluginDir := filepath.Dir(plugin.Command)
			runCommandDetached("open", "-a", "Terminal", pluginDir)
		}
	default:
		if !removed && item != nil {
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

func (pm *PluginMenu) Remove() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.itemId != -1 {
		statusbar.RemoveStatusItem(pm.itemId)
		pm.itemId = -1
	}
	pm.plugin = nil
	pm.menuItems = nil
	pm.prevState = nil
	pm.prevTitle = ""
	pm.prevImageHash = ""
}

func ShowDefault(pluginDir string, onRefreshAll func(), onQuit func()) int {
	itemId := statusbar.CreateStatusItem()
	statusbar.SetTitle(itemId, "xbar")

	statusbar.AddMenuItem(itemId, 0, L("Open Plugin Folder..."), false, false)
	statusbar.AddMenuItem(itemId, 1, L("Refresh All"), false, false)
	statusbar.AddMenuItem(itemId, 2, L("Start at Login"), false, false)
	statusbar.SetMenuItemState(itemId, 2, loginitem.IsEnabled())
	statusbar.AddMenuItem(itemId, -1, "", false, true)
	statusbar.AddMenuItem(itemId, 3, L("Quit"), false, false)

	statusbar.SetClickHandler(func(id int, menuIndex int) {
		if id != itemId {
			return
		}
		switch menuIndex {
		case 0:
			os.MkdirAll(pluginDir, 0755)
			runCommandDetached("open", pluginDir)
		case 1:
			if onRefreshAll != nil {
				go onRefreshAll()
			}
		case 2:
			if err := loginitem.Toggle(); err != nil {
				statusbar.ShowAlert(L("Error"), "Failed to toggle login item: "+err.Error())
			}
			statusbar.SetMenuItemState(itemId, 2, loginitem.IsEnabled())
		case 3:
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

func runCommandDetached(name string, args ...string) {
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		return
	}
	go func() {
		_ = cmd.Wait()
	}()
}
