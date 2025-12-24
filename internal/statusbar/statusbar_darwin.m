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
    NSString *nsTitle = [NSString stringWithUTF8String:title];
    NSString *nsColor = hexColor ? [NSString stringWithUTF8String:hexColor] : nil;

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        XBMenuDelegate *delegate = [menuDelegates objectForKey:@(itemId)];
        if (menu != nil) {
            if (isSeparator) {
                [menu addItem:[NSMenuItem separatorItem]];
            } else {
                NSMenuItem *menuItem = [[NSMenuItem alloc] initWithTitle:nsTitle action:@selector(menuItemClicked:) keyEquivalent:@""];
                [menuItem setTag:menuItemIndex];
                [menuItem setTarget:delegate];
                if (disabled) {
                    [menuItem setEnabled:NO];
                }

                if (nsColor != nil) {
                    NSColor *color = colorFromHex(nsColor);
                    if (color != nil) {
                        NSDictionary *attrs = @{NSForegroundColorAttributeName: color};
                        NSAttributedString *attrTitle = [[NSAttributedString alloc] initWithString:nsTitle attributes:attrs];
                        [menuItem setAttributedTitle:attrTitle];
                    }
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
    NSString *nsSubmenuTitle = [NSString stringWithUTF8String:submenuTitle];
    NSString *nsTitle = [NSString stringWithUTF8String:title];
    NSString *nsColor = hexColor ? [NSString stringWithUTF8String:hexColor] : nil;

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        XBMenuDelegate *delegate = [menuDelegates objectForKey:@(itemId)];
        if (menu != nil) {
            // Find the submenu
            NSMenuItem *parentItem = [menu itemWithTitle:nsSubmenuTitle];
            if (parentItem != nil && [parentItem hasSubmenu]) {
                NSMenu *submenu = [parentItem submenu];

                if (isSeparator) {
                    [submenu addItem:[NSMenuItem separatorItem]];
                } else {
                    NSMenuItem *menuItem = [[NSMenuItem alloc] initWithTitle:nsTitle action:@selector(menuItemClicked:) keyEquivalent:@""];
                    [menuItem setTag:menuItemIndex];
                    [menuItem setTarget:delegate];
                    if (disabled) {
                        [menuItem setEnabled:NO];
                    }

                    if (nsColor != nil) {
                        NSColor *color = colorFromHex(nsColor);
                        if (color != nil) {
                            NSDictionary *attrs = @{NSForegroundColorAttributeName: color};
                            NSAttributedString *attrTitle = [[NSAttributedString alloc] initWithString:nsTitle attributes:attrs];
                            [menuItem setAttributedTitle:attrTitle];
                        }
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
            NSMenuItem *parentItem = [menu itemWithTitle:nsParentTitle];
            if (parentItem != nil && [parentItem hasSubmenu]) {
                NSMenu *parentSubmenu = [parentItem submenu];

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

void setMenuItemIcon(int itemId, int menuItemIndex, const void *data, int length, int isTemplate) {
    NSData *imageData = [NSData dataWithBytes:data length:length];

    dispatch_block_t block = ^{
        NSMenu *menu = [menus objectForKey:@(itemId)];
        if (menu != nil) {
            for (NSMenuItem *item in [menu itemArray]) {
                if ([item tag] == menuItemIndex) {
                    NSImage *image = [[NSImage alloc] initWithData:imageData];
                    if (image != nil) {
                        [image setSize:NSMakeSize(16, 16)];
                        [image setTemplate:(isTemplate != 0)];
                        [item setImage:image];
                    }
                    break;
                }
                // Also check submenus
                if ([item hasSubmenu]) {
                    for (NSMenuItem *subitem in [[item submenu] itemArray]) {
                        if ([subitem tag] == menuItemIndex) {
                            NSImage *image = [[NSImage alloc] initWithData:imageData];
                            if (image != nil) {
                                [image setSize:NSMakeSize(16, 16)];
                                [image setTemplate:(isTemplate != 0)];
                                [subitem setImage:image];
                            }
                            return;
                        }
                        // Check nested submenus too
                        if ([subitem hasSubmenu]) {
                            for (NSMenuItem *nestedItem in [[subitem submenu] itemArray]) {
                                if ([nestedItem tag] == menuItemIndex) {
                                    NSImage *image = [[NSImage alloc] initWithData:imageData];
                                    if (image != nil) {
                                        [image setSize:NSMakeSize(16, 16)];
                                        [image setTemplate:(isTemplate != 0)];
                                        [nestedItem setImage:image];
                                    }
                                    return;
                                }
                            }
                        }
                    }
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
