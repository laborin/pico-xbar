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
void addMenuItemWithColor(int itemId, int menuItemIndex, const char *title, int disabled, int isSeparator, const char *hexColor);
void addSubmenu(int itemId, const char *title);
void addSubmenuItem(int itemId, const char *submenuTitle, int menuItemIndex, const char *title, int disabled, int isSeparator, const char *hexColor);
void addNestedSubmenu(int itemId, const char *parentSubmenuTitle, const char *title);
void setMenuItemIcon(int itemId, int menuItemIndex, const void *data, int length, int isTemplate, int shrink);
void removeStatusItem(int itemId);
void runApp(void);
void stopApp(void);
void copyToClipboard(const char *text);
void showAlert(const char *title, const char *message);
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

	if color == "" {
		C.addMenuItem(C.int(itemId), C.int(menuItemIndex), cTitle, C.int(disabledInt), C.int(separatorInt))
	} else {
		cColor := C.CString(color)
		defer C.free(unsafe.Pointer(cColor))
		C.addMenuItemWithColor(C.int(itemId), C.int(menuItemIndex), cTitle, C.int(disabledInt), C.int(separatorInt), cColor)
	}
}

func AddSubmenu(itemId int, title string) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.addSubmenu(C.int(itemId), cTitle)
}

func AddSubmenuItem(itemId int, submenuTitle string, menuItemIndex int, title string, disabled bool, isSeparator bool, color string) {
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

	var cColor *C.char
	if color != "" {
		cColor = C.CString(color)
		defer C.free(unsafe.Pointer(cColor))
	}

	C.addSubmenuItem(C.int(itemId), cSubmenuTitle, C.int(menuItemIndex), cTitle, C.int(disabledInt), C.int(separatorInt), cColor)
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
