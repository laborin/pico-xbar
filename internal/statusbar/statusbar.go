package statusbar

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#include <stdlib.h>

int createStatusItem(void);
void setTitle(int itemId, const char *title);
void setIcon(int itemId, const void *data, int length, int isTemplate);
void clearMenu(int itemId);
void addMenuItem(int itemId, int menuItemIndex, const char *title, int disabled, int isSeparator);
void addMenuItemStyled(int itemId, int menuItemIndex, const char *title, int disabled, int isSeparator, const char *hexColor, const char *font, int fontSize, const char *keyEquiv, int isAlternate);
void addSubmenu(int itemId, const char *title);
void addSubmenuItemStyled(int itemId, const char *submenuTitle, int menuItemIndex, const char *title, int disabled, int isSeparator, const char *hexColor, const char *font, int fontSize, const char *keyEquiv, int isAlternate);
void addNestedSubmenu(int itemId, const char *parentSubmenuTitle, const char *title);
void setMenuItemIcon(int itemId, int menuItemIndex, const void *data, int length, int isTemplate, int shrink);
int getMenuItemCount(int itemId);
void updateMenuItemAtIndex(int itemId, int atIndex, int menuItemTag, const char *title, int disabled, int isSeparator, const char *hexColor, const char *font, int fontSize, const char *keyEquiv, int isAlternate);
void removeMenuItemAtIndex(int itemId, int atIndex);
void insertMenuItemStyledAtIndex(int itemId, int atIndex, int menuItemTag, const char *title, int disabled, int isSeparator, const char *hexColor, const char *font, int fontSize, const char *keyEquiv, int isAlternate);
void insertSubmenuAtIndex(int itemId, int atIndex, const char *title);
int getSubmenuItemCount(int itemId, const char *submenuTitle);
void updateSubmenuItemAtIndex(int itemId, const char *submenuTitle, int atIndex, int menuItemTag, const char *title, int disabled, int isSeparator, const char *hexColor, const char *font, int fontSize, const char *keyEquiv, int isAlternate);
void removeSubmenuItemAtIndex(int itemId, const char *submenuTitle, int atIndex);
void insertSubmenuItemStyledAtIndex(int itemId, const char *submenuTitle, int atIndex, int menuItemTag, const char *title, int disabled, int isSeparator, const char *hexColor, const char *font, int fontSize, const char *keyEquiv, int isAlternate);
void insertNestedSubmenuAtIndex(int itemId, const char *parentSubmenuTitle, int atIndex, const char *title);
int menuItemIsSubmenu(int itemId, int atIndex);
int submenuItemIsSubmenu(int itemId, const char *submenuTitle, int atIndex);
void removeStatusItem(int itemId);
void runApp(void);
void stopApp(void);
void copyToClipboard(const char *text);
void showAlert(const char *title, const char *message);
void setMenuItemState(int itemId, int menuItemIndex, int state);
*/
import "C"
import (
	"sync"
	"unsafe"
)

type ClickHandler func(itemId int, menuItemIndex int)

var (
	clickHandler ClickHandler
	handlerMu    sync.Mutex
)

//export goClickCallback
func goClickCallback(itemId C.int, menuItemIndex C.int) {
	handlerMu.Lock()
	handler := clickHandler
	handlerMu.Unlock()

	if handler != nil {
		go handler(int(itemId), int(menuItemIndex))
	}
}

func SetClickHandler(handler ClickHandler) {
	handlerMu.Lock()
	clickHandler = handler
	handlerMu.Unlock()
}

func CreateStatusItem() int {
	return int(C.createStatusItem())
}

func SetTitle(itemId int, title string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.setTitle(C.int(itemId), cTitle)
}

func SetIcon(itemId int, data []byte, isTemplate bool) {
	if len(data) == 0 {
		return
	}
	template := 0
	if isTemplate {
		template = 1
	}
	C.setIcon(C.int(itemId), unsafe.Pointer(&data[0]), C.int(len(data)), C.int(template))
}

func ClearMenu(itemId int) {
	C.clearMenu(C.int(itemId))
}

func AddMenuItem(itemId int, menuItemIndex int, title string, disabled bool, isSeparator bool) {
	AddMenuItemWithColor(itemId, menuItemIndex, title, disabled, isSeparator, "")
}

func AddMenuItemWithColor(itemId int, menuItemIndex int, title string, disabled bool, isSeparator bool, color string) {
	AddMenuItemStyled(itemId, menuItemIndex, title, disabled, isSeparator, color, "", 0, "", false)
}

type MenuItemStyle struct {
	Color     string
	Font      string
	Size      int
	Key       string
	Alternate bool
}

func AddMenuItemStyled(itemId int, menuItemIndex int, title string, disabled bool, isSeparator bool, color string, font string, size int, key string, alternate bool) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	disabledInt := 0
	if disabled {
		disabledInt = 1
	}
	separatorInt := 0
	if isSeparator {
		separatorInt = 1
	}
	alternateInt := 0
	if alternate {
		alternateInt = 1
	}

	var cColor *C.char
	if color != "" {
		cColor = C.CString(color)
		defer C.free(unsafe.Pointer(cColor))
	}

	var cFont *C.char
	if font != "" {
		cFont = C.CString(font)
		defer C.free(unsafe.Pointer(cFont))
	}

	var cKey *C.char
	if key != "" {
		cKey = C.CString(key)
		defer C.free(unsafe.Pointer(cKey))
	}

	C.addMenuItemStyled(C.int(itemId), C.int(menuItemIndex), cTitle, C.int(disabledInt), C.int(separatorInt), cColor, cFont, C.int(size), cKey, C.int(alternateInt))
}

func AddSubmenu(itemId int, title string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.addSubmenu(C.int(itemId), cTitle)
}

func AddSubmenuItem(itemId int, submenuTitle string, menuItemIndex int, title string, disabled bool, isSeparator bool, color string) {
	AddSubmenuItemStyled(itemId, submenuTitle, menuItemIndex, title, disabled, isSeparator, color, "", 0, "", false)
}

func AddSubmenuItemStyled(itemId int, submenuTitle string, menuItemIndex int, title string, disabled bool, isSeparator bool, color string, font string, size int, key string, alternate bool) {
	cSubmenuTitle := C.CString(submenuTitle)
	defer C.free(unsafe.Pointer(cSubmenuTitle))
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	disabledInt := 0
	if disabled {
		disabledInt = 1
	}
	separatorInt := 0
	if isSeparator {
		separatorInt = 1
	}
	alternateInt := 0
	if alternate {
		alternateInt = 1
	}

	var cColor *C.char
	if color != "" {
		cColor = C.CString(color)
		defer C.free(unsafe.Pointer(cColor))
	}

	var cFont *C.char
	if font != "" {
		cFont = C.CString(font)
		defer C.free(unsafe.Pointer(cFont))
	}

	var cKey *C.char
	if key != "" {
		cKey = C.CString(key)
		defer C.free(unsafe.Pointer(cKey))
	}

	C.addSubmenuItemStyled(C.int(itemId), cSubmenuTitle, C.int(menuItemIndex), cTitle, C.int(disabledInt), C.int(separatorInt), cColor, cFont, C.int(size), cKey, C.int(alternateInt))
}

func AddNestedSubmenu(itemId int, parentSubmenuTitle string, title string) {
	cParentTitle := C.CString(parentSubmenuTitle)
	defer C.free(unsafe.Pointer(cParentTitle))
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.addNestedSubmenu(C.int(itemId), cParentTitle, cTitle)
}

func SetMenuItemIcon(itemId int, menuItemIndex int, data []byte, isTemplate bool, shrink bool) {
	if len(data) == 0 {
		return
	}
	template := 0
	if isTemplate {
		template = 1
	}
	shrinkInt := 0
	if shrink {
		shrinkInt = 1
	}
	C.setMenuItemIcon(C.int(itemId), C.int(menuItemIndex), unsafe.Pointer(&data[0]), C.int(len(data)), C.int(template), C.int(shrinkInt))
}

func RemoveStatusItem(itemId int) {
	C.removeStatusItem(C.int(itemId))
}

func Run(onReady func()) {
	go onReady()
	C.runApp()
}

func Stop() {
	C.stopApp()
}

func CopyToClipboard(text string) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	C.copyToClipboard(cText)
}

func ShowAlert(title, message string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cMessage))
	C.showAlert(cTitle, cMessage)
}

func SetMenuItemState(itemId int, menuItemIndex int, checked bool) {
	state := 0
	if checked {
		state = 1
	}
	C.setMenuItemState(C.int(itemId), C.int(menuItemIndex), C.int(state))
}

func GetMenuItemCount(itemId int) int {
	return int(C.getMenuItemCount(C.int(itemId)))
}

func UpdateMenuItemAtIndex(itemId int, atIndex int, menuItemTag int, title string, disabled bool, isSeparator bool, color string, font string, size int, key string, alternate bool) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	disabledInt := 0
	if disabled {
		disabledInt = 1
	}
	separatorInt := 0
	if isSeparator {
		separatorInt = 1
	}
	alternateInt := 0
	if alternate {
		alternateInt = 1
	}

	var cColor *C.char
	if color != "" {
		cColor = C.CString(color)
		defer C.free(unsafe.Pointer(cColor))
	}

	var cFont *C.char
	if font != "" {
		cFont = C.CString(font)
		defer C.free(unsafe.Pointer(cFont))
	}

	var cKey *C.char
	if key != "" {
		cKey = C.CString(key)
		defer C.free(unsafe.Pointer(cKey))
	}

	C.updateMenuItemAtIndex(C.int(itemId), C.int(atIndex), C.int(menuItemTag), cTitle, C.int(disabledInt), C.int(separatorInt), cColor, cFont, C.int(size), cKey, C.int(alternateInt))
}

func RemoveMenuItemAtIndex(itemId int, atIndex int) {
	C.removeMenuItemAtIndex(C.int(itemId), C.int(atIndex))
}

func InsertMenuItemStyledAtIndex(itemId int, atIndex int, menuItemTag int, title string, disabled bool, isSeparator bool, color string, font string, size int, key string, alternate bool) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	disabledInt := 0
	if disabled {
		disabledInt = 1
	}
	separatorInt := 0
	if isSeparator {
		separatorInt = 1
	}
	alternateInt := 0
	if alternate {
		alternateInt = 1
	}

	var cColor *C.char
	if color != "" {
		cColor = C.CString(color)
		defer C.free(unsafe.Pointer(cColor))
	}

	var cFont *C.char
	if font != "" {
		cFont = C.CString(font)
		defer C.free(unsafe.Pointer(cFont))
	}

	var cKey *C.char
	if key != "" {
		cKey = C.CString(key)
		defer C.free(unsafe.Pointer(cKey))
	}

	C.insertMenuItemStyledAtIndex(C.int(itemId), C.int(atIndex), C.int(menuItemTag), cTitle, C.int(disabledInt), C.int(separatorInt), cColor, cFont, C.int(size), cKey, C.int(alternateInt))
}

func InsertSubmenuAtIndex(itemId int, atIndex int, title string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.insertSubmenuAtIndex(C.int(itemId), C.int(atIndex), cTitle)
}

func GetSubmenuItemCount(itemId int, submenuTitle string) int {
	cSubmenuTitle := C.CString(submenuTitle)
	defer C.free(unsafe.Pointer(cSubmenuTitle))
	return int(C.getSubmenuItemCount(C.int(itemId), cSubmenuTitle))
}

func UpdateSubmenuItemAtIndex(itemId int, submenuTitle string, atIndex int, menuItemTag int, title string, disabled bool, isSeparator bool, color string, font string, size int, key string, alternate bool) {
	cSubmenuTitle := C.CString(submenuTitle)
	defer C.free(unsafe.Pointer(cSubmenuTitle))
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	disabledInt := 0
	if disabled {
		disabledInt = 1
	}
	separatorInt := 0
	if isSeparator {
		separatorInt = 1
	}
	alternateInt := 0
	if alternate {
		alternateInt = 1
	}

	var cColor *C.char
	if color != "" {
		cColor = C.CString(color)
		defer C.free(unsafe.Pointer(cColor))
	}

	var cFont *C.char
	if font != "" {
		cFont = C.CString(font)
		defer C.free(unsafe.Pointer(cFont))
	}

	var cKey *C.char
	if key != "" {
		cKey = C.CString(key)
		defer C.free(unsafe.Pointer(cKey))
	}

	C.updateSubmenuItemAtIndex(C.int(itemId), cSubmenuTitle, C.int(atIndex), C.int(menuItemTag), cTitle, C.int(disabledInt), C.int(separatorInt), cColor, cFont, C.int(size), cKey, C.int(alternateInt))
}

func RemoveSubmenuItemAtIndex(itemId int, submenuTitle string, atIndex int) {
	cSubmenuTitle := C.CString(submenuTitle)
	defer C.free(unsafe.Pointer(cSubmenuTitle))
	C.removeSubmenuItemAtIndex(C.int(itemId), cSubmenuTitle, C.int(atIndex))
}

func InsertSubmenuItemStyledAtIndex(itemId int, submenuTitle string, atIndex int, menuItemTag int, title string, disabled bool, isSeparator bool, color string, font string, size int, key string, alternate bool) {
	cSubmenuTitle := C.CString(submenuTitle)
	defer C.free(unsafe.Pointer(cSubmenuTitle))
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))

	disabledInt := 0
	if disabled {
		disabledInt = 1
	}
	separatorInt := 0
	if isSeparator {
		separatorInt = 1
	}
	alternateInt := 0
	if alternate {
		alternateInt = 1
	}

	var cColor *C.char
	if color != "" {
		cColor = C.CString(color)
		defer C.free(unsafe.Pointer(cColor))
	}

	var cFont *C.char
	if font != "" {
		cFont = C.CString(font)
		defer C.free(unsafe.Pointer(cFont))
	}

	var cKey *C.char
	if key != "" {
		cKey = C.CString(key)
		defer C.free(unsafe.Pointer(cKey))
	}

	C.insertSubmenuItemStyledAtIndex(C.int(itemId), cSubmenuTitle, C.int(atIndex), C.int(menuItemTag), cTitle, C.int(disabledInt), C.int(separatorInt), cColor, cFont, C.int(size), cKey, C.int(alternateInt))
}

func InsertNestedSubmenuAtIndex(itemId int, parentSubmenuTitle string, atIndex int, title string) {
	cParentTitle := C.CString(parentSubmenuTitle)
	defer C.free(unsafe.Pointer(cParentTitle))
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.insertNestedSubmenuAtIndex(C.int(itemId), cParentTitle, C.int(atIndex), cTitle)
}

func MenuItemIsSubmenu(itemId int, atIndex int) bool {
	return C.menuItemIsSubmenu(C.int(itemId), C.int(atIndex)) != 0
}

func SubmenuItemIsSubmenu(itemId int, submenuTitle string, atIndex int) bool {
	cSubmenuTitle := C.CString(submenuTitle)
	defer C.free(unsafe.Pointer(cSubmenuTitle))
	return C.submenuItemIsSubmenu(C.int(itemId), cSubmenuTitle, C.int(atIndex)) != 0
}
