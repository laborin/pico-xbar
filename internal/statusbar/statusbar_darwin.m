#import <Cocoa/Cocoa.h>
#import <dispatch/dispatch.h>

static NSMutableDictionary *statusItems = nil;
static NSMutableDictionary *menus = nil;
static int nextItemId = 0;

extern void goClickCallback(int itemId, int menuItemIndex);

@interface XBMenuDelegate : NSObject
@property (nonatomic, assign) int itemId;
- (void)menuItemClicked:(id)sender;
@end

@implementation XBMenuDelegate
- (void)menuItemClicked:(id)sender {
    goClickCallback(self.itemId, (int)[(NSMenuItem *)sender tag]);
}
@end

static NSMutableDictionary *menuDelegates = nil;

// Forward declarations
void addMenuItemWithColor(int itemId, int menuItemIndex, const char *title, int disabled, int isSeparator, const char *hexColor);
void addMenuItemStyled(int itemId, int menuItemIndex, const char *title, int disabled, int isSeparator, const char *hexColor, const char *font, int fontSize, const char *keyEquiv, int isAlternate);
void addSubmenuItem(int itemId, const char *submenuTitle, int menuItemIndex, const char *title, int disabled, int isSeparator, const char *hexColor);
void addSubmenuItemStyled(int itemId, const char *submenuTitle, int menuItemIndex, const char *title, int disabled, int isSeparator, const char *hexColor, const char *font, int fontSize, const char *keyEquiv, int isAlternate);

static NSMenu* findSubmenuByTitle(NSMenu *menu, NSString *title) {
    for (NSMenuItem *item in [menu itemArray]) {
        if ([item hasSubmenu]) {
            if ([[item title] isEqualToString:title]) {
                return [item submenu];
            }
            NSMenu *found = findSubmenuByTitle([item submenu], title);
            if (found != nil) {
                return found;
            }
        }
    }
    return nil;
}

static NSColor* colorFromHex(NSString *hexString) {
    if (hexString == nil || [hexString length] == 0) {
        return nil;
    }

    hexString = [hexString stringByReplacingOccurrencesOfString:@"#" withString:@""];

    unsigned int r, g, b;
    float a = 1.0;

    if ([hexString length] == 3) {
        // #RGB
        [[NSScanner scannerWithString:[hexString substringWithRange:NSMakeRange(0, 1)]] scanHexInt:&r];
        [[NSScanner scannerWithString:[hexString substringWithRange:NSMakeRange(1, 1)]] scanHexInt:&g];
        [[NSScanner scannerWithString:[hexString substringWithRange:NSMakeRange(2, 1)]] scanHexInt:&b];
        r *= 17; g *= 17; b *= 17;
    } else if ([hexString length] == 4) {
        // #RGBA
        [[NSScanner scannerWithString:[hexString substringWithRange:NSMakeRange(0, 1)]] scanHexInt:&r];
        [[NSScanner scannerWithString:[hexString substringWithRange:NSMakeRange(1, 1)]] scanHexInt:&g];
        [[NSScanner scannerWithString:[hexString substringWithRange:NSMakeRange(2, 1)]] scanHexInt:&b];
        unsigned int aInt;
        [[NSScanner scannerWithString:[hexString substringWithRange:NSMakeRange(3, 1)]] scanHexInt:&aInt];
        r *= 17; g *= 17; b *= 17;
        a = (float)(aInt * 17) / 255.0;
    } else if ([hexString length] == 6) {
        // #RRGGBB
        [[NSScanner scannerWithString:[hexString substringWithRange:NSMakeRange(0, 2)]] scanHexInt:&r];
        [[NSScanner scannerWithString:[hexString substringWithRange:NSMakeRange(2, 2)]] scanHexInt:&g];
        [[NSScanner scannerWithString:[hexString substringWithRange:NSMakeRange(4, 2)]] scanHexInt:&b];
    } else if ([hexString length] == 8) {
        // #RRGGBBAA
        [[NSScanner scannerWithString:[hexString substringWithRange:NSMakeRange(0, 2)]] scanHexInt:&r];
        [[NSScanner scannerWithString:[hexString substringWithRange:NSMakeRange(2, 2)]] scanHexInt:&g];
        [[NSScanner scannerWithString:[hexString substringWithRange:NSMakeRange(4, 2)]] scanHexInt:&b];
        unsigned int aInt;
        [[NSScanner scannerWithString:[hexString substringWithRange:NSMakeRange(6, 2)]] scanHexInt:&aInt];
        a = (float)aInt / 255.0;
    } else {
        return nil;
    }

    return [NSColor colorWithCalibratedRed:(float)r/255.0 green:(float)g/255.0 blue:(float)b/255.0 alpha:a];
}

int createStatusItem(void) {
    __block int itemId;

    dispatch_block_t block = ^{
        if (statusItems == nil) {
            statusItems = [[NSMutableDictionary alloc] init];
            menus = [[NSMutableDictionary alloc] init];
            menuDelegates = [[NSMutableDictionary alloc] init];
        }

        itemId = nextItemId++;

        NSStatusItem *item = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
        item.button.toolTip = @"xbar-light";

        NSMenu *menu = [[NSMenu alloc] init];
        [menu setAutoenablesItems:NO];
        [item setMenu:menu];

        XBMenuDelegate *delegate = [[XBMenuDelegate alloc] init];
        delegate.itemId = itemId;

        [statusItems setObject:item forKey:@(itemId)];
        [menus setObject:menu forKey:@(itemId)];
        [menuDelegates setObject:delegate forKey:@(itemId)];
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_sync(dispatch_get_main_queue(), block);
    }

    return itemId;
}

void setTitle(int itemId, const char *title) {
    NSString *nsTitle = [NSString stringWithUTF8String:title];

    dispatch_block_t block = ^{
        NSStatusItem *item = [statusItems objectForKey:@(itemId)];
        if (item != nil) {
            [item.button setTitle:nsTitle];
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void setIcon(int itemId, const void *data, int length, int isTemplate) {
    NSData *imageData = [NSData dataWithBytes:data length:length];

    dispatch_block_t block = ^{
        NSStatusItem *item = [statusItems objectForKey:@(itemId)];
        if (item != nil) {
            NSImage *image = [[NSImage alloc] initWithData:imageData];
            if (image != nil) {
                [image setSize:NSMakeSize(18, 18)];
                [image setTemplate:(isTemplate != 0)];
                [item.button setImage:image];
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void clearMenu(int itemId) {
    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        if (menu != nil) {
            [menu removeAllItems];
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void addMenuItem(int itemId, int menuItemIndex, const char *title, int disabled, int isSeparator) {
    addMenuItemWithColor(itemId, menuItemIndex, title, disabled, isSeparator, NULL);
}

void addMenuItemWithColor(int itemId, int menuItemIndex, const char *title, int disabled, int isSeparator, const char *hexColor) {
    addMenuItemStyled(itemId, menuItemIndex, title, disabled, isSeparator, hexColor, NULL, 0, NULL, 0);
}

void addMenuItemStyled(int itemId, int menuItemIndex, const char *title, int disabled, int isSeparator, const char *hexColor, const char *font, int fontSize, const char *keyEquiv, int isAlternate) {
    NSString *nsTitle = [NSString stringWithUTF8String:title];
    NSString *nsColor = hexColor ? [NSString stringWithUTF8String:hexColor] : nil;
    NSString *nsFont = font ? [NSString stringWithUTF8String:font] : nil;
    NSString *nsKeyEquiv = keyEquiv ? [NSString stringWithUTF8String:keyEquiv] : @"";

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        XBMenuDelegate *delegate = [menuDelegates objectForKey:@(itemId)];
        if (menu != nil) {
            if (isSeparator) {
                [menu addItem:[NSMenuItem separatorItem]];
            } else {
                NSMenuItem *menuItem = [[NSMenuItem alloc] initWithTitle:nsTitle action:@selector(menuItemClicked:) keyEquivalent:nsKeyEquiv];
                [menuItem setTag:menuItemIndex];
                [menuItem setTarget:delegate];
                if (disabled) {
                    [menuItem setEnabled:NO];
                }
                if (isAlternate) {
                    [menuItem setAlternate:YES];
                    [menuItem setKeyEquivalentModifierMask:NSEventModifierFlagOption];
                }

                NSMutableDictionary *attrs = [NSMutableDictionary dictionary];
                if (nsColor != nil) {
                    NSColor *color = colorFromHex(nsColor);
                    if (color != nil) {
                        [attrs setObject:color forKey:NSForegroundColorAttributeName];
                    }
                }
                if (nsFont != nil || fontSize > 0) {
                    NSFont *fontObj = nil;
                    CGFloat size = fontSize > 0 ? (CGFloat)fontSize : [NSFont systemFontSize];
                    if (nsFont != nil) {
                        fontObj = [NSFont fontWithName:nsFont size:size];
                    }
                    if (fontObj == nil) {
                        fontObj = [NSFont systemFontOfSize:size];
                    }
                    [attrs setObject:fontObj forKey:NSFontAttributeName];
                }
                if ([attrs count] > 0) {
                    NSAttributedString *attrTitle = [[NSAttributedString alloc] initWithString:nsTitle attributes:attrs];
                    [menuItem setAttributedTitle:attrTitle];
                }

                [menu addItem:menuItem];
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void addSubmenu(int itemId, const char *title) {
    NSString *nsTitle = [NSString stringWithUTF8String:title];

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        if (menu != nil) {
            NSMenuItem *menuItem = [[NSMenuItem alloc] initWithTitle:nsTitle action:nil keyEquivalent:@""];
            NSMenu *submenu = [[NSMenu alloc] initWithTitle:nsTitle];
            [submenu setAutoenablesItems:NO];
            [menuItem setSubmenu:submenu];
            [menu addItem:menuItem];
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void addSubmenuItem(int itemId, const char *submenuTitle, int menuItemIndex, const char *title, int disabled, int isSeparator, const char *hexColor) {
    addSubmenuItemStyled(itemId, submenuTitle, menuItemIndex, title, disabled, isSeparator, hexColor, NULL, 0, NULL, 0);
}

void addSubmenuItemStyled(int itemId, const char *submenuTitle, int menuItemIndex, const char *title, int disabled, int isSeparator, const char *hexColor, const char *font, int fontSize, const char *keyEquiv, int isAlternate) {
    NSString *nsSubmenuTitle = [NSString stringWithUTF8String:submenuTitle];
    NSString *nsTitle = [NSString stringWithUTF8String:title];
    NSString *nsColor = hexColor ? [NSString stringWithUTF8String:hexColor] : nil;
    NSString *nsFont = font ? [NSString stringWithUTF8String:font] : nil;
    NSString *nsKeyEquiv = keyEquiv ? [NSString stringWithUTF8String:keyEquiv] : @"";

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        XBMenuDelegate *delegate = [menuDelegates objectForKey:@(itemId)];
        if (menu != nil) {
            NSMenu *submenu = findSubmenuByTitle(menu, nsSubmenuTitle);
            if (submenu != nil) {
                if (isSeparator) {
                    [submenu addItem:[NSMenuItem separatorItem]];
                } else {
                    NSMenuItem *menuItem = [[NSMenuItem alloc] initWithTitle:nsTitle action:@selector(menuItemClicked:) keyEquivalent:nsKeyEquiv];
                    [menuItem setTag:menuItemIndex];
                    [menuItem setTarget:delegate];
                    if (disabled) {
                        [menuItem setEnabled:NO];
                    }
                    if (isAlternate) {
                        [menuItem setAlternate:YES];
                        [menuItem setKeyEquivalentModifierMask:NSEventModifierFlagOption];
                    }

                    NSMutableDictionary *attrs = [NSMutableDictionary dictionary];
                    if (nsColor != nil) {
                        NSColor *color = colorFromHex(nsColor);
                        if (color != nil) {
                            [attrs setObject:color forKey:NSForegroundColorAttributeName];
                        }
                    }
                    if (nsFont != nil || fontSize > 0) {
                        NSFont *fontObj = nil;
                        CGFloat size = fontSize > 0 ? (CGFloat)fontSize : [NSFont systemFontSize];
                        if (nsFont != nil) {
                            fontObj = [NSFont fontWithName:nsFont size:size];
                        }
                        if (fontObj == nil) {
                            fontObj = [NSFont systemFontOfSize:size];
                        }
                        [attrs setObject:fontObj forKey:NSFontAttributeName];
                    }
                    if ([attrs count] > 0) {
                        NSAttributedString *attrTitle = [[NSAttributedString alloc] initWithString:nsTitle attributes:attrs];
                        [menuItem setAttributedTitle:attrTitle];
                    }

                    [submenu addItem:menuItem];
                }
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void addNestedSubmenu(int itemId, const char *parentSubmenuTitle, const char *title) {
    NSString *nsParentTitle = [NSString stringWithUTF8String:parentSubmenuTitle];
    NSString *nsTitle = [NSString stringWithUTF8String:title];

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        if (menu != nil) {
            NSMenu *parentSubmenu = findSubmenuByTitle(menu, nsParentTitle);
            if (parentSubmenu != nil) {
                NSMenuItem *menuItem = [[NSMenuItem alloc] initWithTitle:nsTitle action:nil keyEquivalent:@""];
                NSMenu *submenu = [[NSMenu alloc] initWithTitle:nsTitle];
                [submenu setAutoenablesItems:NO];
                [menuItem setSubmenu:submenu];
                [parentSubmenu addItem:menuItem];
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

static NSMenuItem* findMenuItemByTag(NSMenu *menu, int tag) {
    for (NSMenuItem *item in [menu itemArray]) {
        if ([item tag] == tag && ![item hasSubmenu]) {
            return item;
        }
        if ([item hasSubmenu]) {
            NSMenuItem *found = findMenuItemByTag([item submenu], tag);
            if (found != nil) {
                return found;
            }
        }
    }
    return nil;
}

void setMenuItemIcon(int itemId, int menuItemIndex, const void *data, int length, int isTemplate, int shrink) {
    NSData *imageData = [NSData dataWithBytes:data length:length];

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        if (menu != nil) {
            NSMenuItem *menuItem = findMenuItemByTag(menu, menuItemIndex);
            if (menuItem != nil) {
                NSImage *image = [[NSImage alloc] initWithData:imageData];
                if (image != nil) {
                    if (shrink != 0) {
                        [image setSize:NSMakeSize(16, 16)];
                    }
                    [image setTemplate:(isTemplate != 0)];
                    [menuItem setImage:image];
                }
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void removeStatusItem(int itemId) {
    dispatch_block_t block = ^{
        NSStatusItem *item = [statusItems objectForKey:@(itemId)];
        if (item != nil) {
            [[NSStatusBar systemStatusBar] removeStatusItem:item];
            [statusItems removeObjectForKey:@(itemId)];
            [menus removeObjectForKey:@(itemId)];
            [menuDelegates removeObjectForKey:@(itemId)];
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void runApp(void) {
    @autoreleasepool {
        [NSApplication sharedApplication];
        [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
        [NSApp run];
    }
}

void stopApp(void) {
    dispatch_block_t block = ^{
        [NSApp terminate:nil];
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void copyToClipboard(const char *text) {
    NSString *nsText = [NSString stringWithUTF8String:text];

    dispatch_block_t block = ^{
        NSPasteboard *pasteboard = [NSPasteboard generalPasteboard];
        [pasteboard clearContents];
        [pasteboard setString:nsText forType:NSPasteboardTypeString];
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void showAlert(const char *title, const char *message) {
    NSString *nsTitle = [NSString stringWithUTF8String:title];
    NSString *nsMessage = [NSString stringWithUTF8String:message];

    dispatch_block_t block = ^{
        NSAlert *alert = [[NSAlert alloc] init];
        [alert setMessageText:nsTitle];
        [alert setInformativeText:nsMessage];
        [alert addButtonWithTitle:@"OK"];
        [alert runModal];
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void setMenuItemState(int itemId, int menuItemIndex, int state) {
    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        if (menu != nil) {
            NSMenuItem *menuItem = findMenuItemByTag(menu, menuItemIndex);
            if (menuItem != nil) {
                [menuItem setState:(state != 0) ? NSControlStateValueOn : NSControlStateValueOff];
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

int getMenuItemCount(int itemId) {
    __block int count = 0;

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        if (menu != nil) {
            count = (int)[[menu itemArray] count];
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_sync(dispatch_get_main_queue(), block);
    }

    return count;
}

static void applyStylesToMenuItem(NSMenuItem *menuItem, NSString *nsTitle, NSString *nsColor, NSString *nsFont, int fontSize) {
    NSMutableDictionary *attrs = [NSMutableDictionary dictionary];
    if (nsColor != nil) {
        NSColor *color = colorFromHex(nsColor);
        if (color != nil) {
            [attrs setObject:color forKey:NSForegroundColorAttributeName];
        }
    }
    if (nsFont != nil || fontSize > 0) {
        NSFont *fontObj = nil;
        CGFloat size = fontSize > 0 ? (CGFloat)fontSize : [NSFont systemFontSize];
        if (nsFont != nil) {
            fontObj = [NSFont fontWithName:nsFont size:size];
        }
        if (fontObj == nil) {
            fontObj = [NSFont systemFontOfSize:size];
        }
        [attrs setObject:fontObj forKey:NSFontAttributeName];
    }
    if ([attrs count] > 0) {
        NSAttributedString *attrTitle = [[NSAttributedString alloc] initWithString:nsTitle attributes:attrs];
        [menuItem setAttributedTitle:attrTitle];
    } else {
        [menuItem setTitle:nsTitle];
        [menuItem setAttributedTitle:nil];
    }
}

void updateMenuItemAtIndex(int itemId, int atIndex, int menuItemTag, const char *title, int disabled, int isSeparator, const char *hexColor, const char *font, int fontSize, const char *keyEquiv, int isAlternate) {
    NSString *nsTitle = title ? [NSString stringWithUTF8String:title] : @"";
    NSString *nsColor = hexColor ? [NSString stringWithUTF8String:hexColor] : nil;
    NSString *nsFont = font ? [NSString stringWithUTF8String:font] : nil;
    NSString *nsKeyEquiv = keyEquiv ? [NSString stringWithUTF8String:keyEquiv] : @"";

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        XBMenuDelegate *delegate = [menuDelegates objectForKey:@(itemId)];
        if (menu != nil && atIndex < [[menu itemArray] count]) {
            NSMenuItem *menuItem = [menu itemAtIndex:atIndex];
            if (menuItem != nil && ![menuItem hasSubmenu]) {
                [menuItem setTag:menuItemTag];
                [menuItem setTarget:delegate];
                [menuItem setAction:@selector(menuItemClicked:)];
                [menuItem setEnabled:!disabled];
                [menuItem setKeyEquivalent:nsKeyEquiv];
                [menuItem setAlternate:(isAlternate != 0)];
                if (isAlternate) {
                    [menuItem setKeyEquivalentModifierMask:NSEventModifierFlagOption];
                } else {
                    [menuItem setKeyEquivalentModifierMask:0];
                }
                applyStylesToMenuItem(menuItem, nsTitle, nsColor, nsFont, fontSize);
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void removeMenuItemAtIndex(int itemId, int atIndex) {
    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        if (menu != nil && atIndex < [[menu itemArray] count]) {
            [menu removeItemAtIndex:atIndex];
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void insertMenuItemStyledAtIndex(int itemId, int atIndex, int menuItemTag, const char *title, int disabled, int isSeparator, const char *hexColor, const char *font, int fontSize, const char *keyEquiv, int isAlternate) {
    NSString *nsTitle = title ? [NSString stringWithUTF8String:title] : @"";
    NSString *nsColor = hexColor ? [NSString stringWithUTF8String:hexColor] : nil;
    NSString *nsFont = font ? [NSString stringWithUTF8String:font] : nil;
    NSString *nsKeyEquiv = keyEquiv ? [NSString stringWithUTF8String:keyEquiv] : @"";

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        XBMenuDelegate *delegate = [menuDelegates objectForKey:@(itemId)];
        if (menu != nil) {
            NSMenuItem *menuItem;
            if (isSeparator) {
                menuItem = [NSMenuItem separatorItem];
            } else {
                menuItem = [[NSMenuItem alloc] initWithTitle:nsTitle action:@selector(menuItemClicked:) keyEquivalent:nsKeyEquiv];
                [menuItem setTag:menuItemTag];
                [menuItem setTarget:delegate];
                if (disabled) {
                    [menuItem setEnabled:NO];
                }
                if (isAlternate) {
                    [menuItem setAlternate:YES];
                    [menuItem setKeyEquivalentModifierMask:NSEventModifierFlagOption];
                }
                applyStylesToMenuItem(menuItem, nsTitle, nsColor, nsFont, fontSize);
            }
            if (atIndex >= [[menu itemArray] count]) {
                [menu addItem:menuItem];
            } else {
                [menu insertItem:menuItem atIndex:atIndex];
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void insertSubmenuAtIndex(int itemId, int atIndex, const char *title) {
    NSString *nsTitle = [NSString stringWithUTF8String:title];

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        if (menu != nil) {
            NSMenuItem *menuItem = [[NSMenuItem alloc] initWithTitle:nsTitle action:nil keyEquivalent:@""];
            NSMenu *submenu = [[NSMenu alloc] initWithTitle:nsTitle];
            [submenu setAutoenablesItems:NO];
            [menuItem setSubmenu:submenu];
            if (atIndex >= [[menu itemArray] count]) {
                [menu addItem:menuItem];
            } else {
                [menu insertItem:menuItem atIndex:atIndex];
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

int getSubmenuItemCount(int itemId, const char *submenuTitle) {
    __block int count = 0;
    NSString *nsSubmenuTitle = [NSString stringWithUTF8String:submenuTitle];

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        if (menu != nil) {
            NSMenu *submenu = findSubmenuByTitle(menu, nsSubmenuTitle);
            if (submenu != nil) {
                count = (int)[[submenu itemArray] count];
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_sync(dispatch_get_main_queue(), block);
    }

    return count;
}

void updateSubmenuItemAtIndex(int itemId, const char *submenuTitle, int atIndex, int menuItemTag, const char *title, int disabled, int isSeparator, const char *hexColor, const char *font, int fontSize, const char *keyEquiv, int isAlternate) {
    NSString *nsSubmenuTitle = [NSString stringWithUTF8String:submenuTitle];
    NSString *nsTitle = title ? [NSString stringWithUTF8String:title] : @"";
    NSString *nsColor = hexColor ? [NSString stringWithUTF8String:hexColor] : nil;
    NSString *nsFont = font ? [NSString stringWithUTF8String:font] : nil;
    NSString *nsKeyEquiv = keyEquiv ? [NSString stringWithUTF8String:keyEquiv] : @"";

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        XBMenuDelegate *delegate = [menuDelegates objectForKey:@(itemId)];
        if (menu != nil) {
            NSMenu *submenu = findSubmenuByTitle(menu, nsSubmenuTitle);
            if (submenu != nil && atIndex < [[submenu itemArray] count]) {
                NSMenuItem *menuItem = [submenu itemAtIndex:atIndex];
                if (menuItem != nil && ![menuItem hasSubmenu]) {
                    [menuItem setTag:menuItemTag];
                    [menuItem setTarget:delegate];
                    [menuItem setAction:@selector(menuItemClicked:)];
                    [menuItem setEnabled:!disabled];
                    [menuItem setKeyEquivalent:nsKeyEquiv];
                    [menuItem setAlternate:(isAlternate != 0)];
                    if (isAlternate) {
                        [menuItem setKeyEquivalentModifierMask:NSEventModifierFlagOption];
                    } else {
                        [menuItem setKeyEquivalentModifierMask:0];
                    }
                    applyStylesToMenuItem(menuItem, nsTitle, nsColor, nsFont, fontSize);
                }
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void removeSubmenuItemAtIndex(int itemId, const char *submenuTitle, int atIndex) {
    NSString *nsSubmenuTitle = [NSString stringWithUTF8String:submenuTitle];

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        if (menu != nil) {
            NSMenu *submenu = findSubmenuByTitle(menu, nsSubmenuTitle);
            if (submenu != nil && atIndex < [[submenu itemArray] count]) {
                [submenu removeItemAtIndex:atIndex];
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void insertSubmenuItemStyledAtIndex(int itemId, const char *submenuTitle, int atIndex, int menuItemTag, const char *title, int disabled, int isSeparator, const char *hexColor, const char *font, int fontSize, const char *keyEquiv, int isAlternate) {
    NSString *nsSubmenuTitle = [NSString stringWithUTF8String:submenuTitle];
    NSString *nsTitle = title ? [NSString stringWithUTF8String:title] : @"";
    NSString *nsColor = hexColor ? [NSString stringWithUTF8String:hexColor] : nil;
    NSString *nsFont = font ? [NSString stringWithUTF8String:font] : nil;
    NSString *nsKeyEquiv = keyEquiv ? [NSString stringWithUTF8String:keyEquiv] : @"";

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        XBMenuDelegate *delegate = [menuDelegates objectForKey:@(itemId)];
        if (menu != nil) {
            NSMenu *submenu = findSubmenuByTitle(menu, nsSubmenuTitle);
            if (submenu != nil) {
                NSMenuItem *menuItem;
                if (isSeparator) {
                    menuItem = [NSMenuItem separatorItem];
                } else {
                    menuItem = [[NSMenuItem alloc] initWithTitle:nsTitle action:@selector(menuItemClicked:) keyEquivalent:nsKeyEquiv];
                    [menuItem setTag:menuItemTag];
                    [menuItem setTarget:delegate];
                    if (disabled) {
                        [menuItem setEnabled:NO];
                    }
                    if (isAlternate) {
                        [menuItem setAlternate:YES];
                        [menuItem setKeyEquivalentModifierMask:NSEventModifierFlagOption];
                    }
                    applyStylesToMenuItem(menuItem, nsTitle, nsColor, nsFont, fontSize);
                }
                if (atIndex >= [[submenu itemArray] count]) {
                    [submenu addItem:menuItem];
                } else {
                    [submenu insertItem:menuItem atIndex:atIndex];
                }
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

void insertNestedSubmenuAtIndex(int itemId, const char *parentSubmenuTitle, int atIndex, const char *title) {
    NSString *nsParentTitle = [NSString stringWithUTF8String:parentSubmenuTitle];
    NSString *nsTitle = [NSString stringWithUTF8String:title];

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        if (menu != nil) {
            NSMenu *parentSubmenu = findSubmenuByTitle(menu, nsParentTitle);
            if (parentSubmenu != nil) {
                NSMenuItem *menuItem = [[NSMenuItem alloc] initWithTitle:nsTitle action:nil keyEquivalent:@""];
                NSMenu *submenu = [[NSMenu alloc] initWithTitle:nsTitle];
                [submenu setAutoenablesItems:NO];
                [menuItem setSubmenu:submenu];
                if (atIndex >= [[parentSubmenu itemArray] count]) {
                    [parentSubmenu addItem:menuItem];
                } else {
                    [parentSubmenu insertItem:menuItem atIndex:atIndex];
                }
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_async(dispatch_get_main_queue(), block);
    }
}

int menuItemIsSubmenu(int itemId, int atIndex) {
    __block int result = 0;

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        if (menu != nil && atIndex < [[menu itemArray] count]) {
            NSMenuItem *item = [menu itemAtIndex:atIndex];
            if ([item hasSubmenu]) {
                result = 1;
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_sync(dispatch_get_main_queue(), block);
    }

    return result;
}

int submenuItemIsSubmenu(int itemId, const char *submenuTitle, int atIndex) {
    __block int result = 0;
    NSString *nsSubmenuTitle = [NSString stringWithUTF8String:submenuTitle];

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        if (menu != nil) {
            NSMenu *submenu = findSubmenuByTitle(menu, nsSubmenuTitle);
            if (submenu != nil && atIndex < [[submenu itemArray] count]) {
                NSMenuItem *item = [submenu itemAtIndex:atIndex];
                if ([item hasSubmenu]) {
                    result = 1;
                }
            }
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_sync(dispatch_get_main_queue(), block);
    }

    return result;
}
