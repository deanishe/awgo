//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

package workflow

import "fmt"

// Ready-to-use icons based on built-in OS X system icons.
// These icons are all found in
// /System/Library/CoreServices/CoreTypes.bundle/Contents/Resources.
//
// The icons are the same as found in the Alfred-Workflow library
// for Python. Preview them here:
// http://www.deanishe.net/alfred-workflow/user-manual/icons.html#list-of-icons
var (
	// Accounts.icns
	IconAccount *Icon
	// BurningIcon.icns
	IconBurn *Icon
	// Clock.icns
	IconClock *Icon
	// ProfileBackgroundColor.icns
	IconColor *Icon
	// ProfileBackgroundColor.icns
	IconColour *Icon
	// EjectMediaIcon.icns
	IconEject *Icon
	// AlertStopIcon.icns
	IconError *Icon
	// ToolbarFavoritesIcon.icns
	IconFavorite *Icon
	// ToolbarFavoritesIcon.icns
	IconFavourite *Icon
	// GroupIcon.icns
	IconGroup *Icon
	// HelpIcon.icns
	IconHelp *Icon
	// HomeFolderIcon.icns
	IconHome *Icon
	// ToolbarInfo.icns
	IconInfo *Icon
	// GenericNetworkIcon.icns
	IconNetwork *Icon
	// AlertNoteIcon.icns
	IconNote *Icon
	// ToolbarAdvanced.icns
	IconSettings *Icon
	// ErasingIcon.icns
	IconSwirl *Icon
	// General.icns
	IconSwitch *Icon
	// Sync.icns
	IconSync *Icon
	// TrashIcon.icns
	IconTrash *Icon
	// UserIcon.icns
	IconUser *Icon
	// AlertCautionIcon.icns
	IconWarning *Icon
	// BookmarkIcon.icns
	IconWeb *Icon
)

func systemIcon(filename string) *Icon {
	icon := &Icon{}
	var path string
	path = fmt.Sprintf(
		"/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/%s.icns", filename)
	icon.Value = path
	icon.Type = ""
	return icon
}

func init() {
	IconAccount = systemIcon("Accounts")
	IconBurn = systemIcon("BurningIcon")
	IconClock = systemIcon("Clock")
	IconColor = systemIcon("ProfileBackgroundColor")
	IconColour = systemIcon("ProfileBackgroundColor")
	IconEject = systemIcon("EjectMediaIcon")
	IconError = systemIcon("AlertStopIcon")
	IconFavorite = systemIcon("ToolbarFavoritesIcon")
	IconFavourite = systemIcon("ToolbarFavoritesIcon")
	IconGroup = systemIcon("GroupIcon")
	IconHelp = systemIcon("HelpIcon")
	IconHome = systemIcon("HomeFolderIcon")
	IconInfo = systemIcon("ToolbarInfo")
	IconNetwork = systemIcon("GenericNetworkIcon")
	IconNote = systemIcon("AlertNoteIcon")
	IconSettings = systemIcon("ToolbarAdvanced")
	IconSwirl = systemIcon("ErasingIcon")
	IconSwitch = systemIcon("General")
	IconSync = systemIcon("Sync")
	IconTrash = systemIcon("TrashIcon")
	IconUser = systemIcon("UserIcon")
	IconWarning = systemIcon("AlertCautionIcon")
	IconWeb = systemIcon("BookmarkIcon")
}
