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
	IconAccount *ItemIcon
	// BurningIcon.icns
	IconBurn *ItemIcon
	// Clock.icns
	IconClock *ItemIcon
	// ProfileBackgroundColor.icns
	IconColor *ItemIcon
	// ProfileBackgroundColor.icns
	IconColour *ItemIcon
	// EjectMediaIcon.icns
	IconEject *ItemIcon
	// AlertStopIcon.icns
	IconError *ItemIcon
	// ToolbarFavoritesIcon.icns
	IconFavorite *ItemIcon
	// ToolbarFavoritesIcon.icns
	IconFavourite *ItemIcon
	// GroupIcon.icns
	IconGroup *ItemIcon
	// HelpIcon.icns
	IconHelp *ItemIcon
	// HomeFolderIcon.icns
	IconHome *ItemIcon
	// ToolbarInfo.icns
	IconInfo *ItemIcon
	// GenericNetworkIcon.icns
	IconNetwork *ItemIcon
	// AlertNoteIcon.icns
	IconNote *ItemIcon
	// ToolbarAdvanced.icns
	IconSettings *ItemIcon
	// ErasingIcon.icns
	IconSwirl *ItemIcon
	// General.icns
	IconSwitch *ItemIcon
	// Sync.icns
	IconSync *ItemIcon
	// TrashIcon.icns
	IconTrash *ItemIcon
	// UserIcon.icns
	IconUser *ItemIcon
	// AlertCautionIcon.icns
	IconWarning *ItemIcon
	// BookmarkIcon.icns
	IconWeb *ItemIcon
)

func systemIcon(filename string) *ItemIcon {
	icon := &ItemIcon{}
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
