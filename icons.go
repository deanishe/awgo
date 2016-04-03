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
	ICON_ACCOUNT *ItemIcon
	// BurningIcon.icns
	ICON_BURN *ItemIcon
	// Clock.icns
	ICON_CLOCK *ItemIcon
	// ProfileBackgroundColor.icns
	ICON_COLOR *ItemIcon
	// ProfileBackgroundColor.icns
	ICON_COLOUR *ItemIcon
	// EjectMediaIcon.icns
	ICON_EJECT *ItemIcon
	// AlertStopIcon.icns
	ICON_ERROR *ItemIcon
	// ToolbarFavoritesIcon.icns
	ICON_FAVORITE *ItemIcon
	// ToolbarFavoritesIcon.icns
	ICON_FAVOURITE *ItemIcon
	// GroupIcon.icns
	ICON_GROUP *ItemIcon
	// HelpIcon.icns
	ICON_HELP *ItemIcon
	// HomeFolderIcon.icns
	ICON_HOME *ItemIcon
	// ToolbarInfo.icns
	ICON_INFO *ItemIcon
	// GenericNetworkIcon.icns
	ICON_NETWORK *ItemIcon
	// AlertNoteIcon.icns
	ICON_NOTE *ItemIcon
	// ToolbarAdvanced.icns
	ICON_SETTINGS *ItemIcon
	// ErasingIcon.icns
	ICON_SWIRL *ItemIcon
	// General.icns
	ICON_SWITCH *ItemIcon
	// Sync.icns
	ICON_SYNC *ItemIcon
	// TrashIcon.icns
	ICON_TRASH *ItemIcon
	// UserIcon.icns
	ICON_USER *ItemIcon
	// AlertCautionIcon.icns
	ICON_WARNING *ItemIcon
	// BookmarkIcon.icns
	ICON_WEB *ItemIcon
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
	ICON_ACCOUNT = systemIcon("Accounts")
	ICON_BURN = systemIcon("BurningIcon")
	ICON_CLOCK = systemIcon("Clock")
	ICON_COLOR = systemIcon("ProfileBackgroundColor")
	ICON_COLOUR = systemIcon("ProfileBackgroundColor")
	ICON_EJECT = systemIcon("EjectMediaIcon")
	ICON_ERROR = systemIcon("AlertStopIcon")
	ICON_FAVORITE = systemIcon("ToolbarFavoritesIcon")
	ICON_FAVOURITE = systemIcon("ToolbarFavoritesIcon")
	ICON_GROUP = systemIcon("GroupIcon")
	ICON_HELP = systemIcon("HelpIcon")
	ICON_HOME = systemIcon("HomeFolderIcon")
	ICON_INFO = systemIcon("ToolbarInfo")
	ICON_NETWORK = systemIcon("GenericNetworkIcon")
	ICON_NOTE = systemIcon("AlertNoteIcon")
	ICON_SETTINGS = systemIcon("ToolbarAdvanced")
	ICON_SWIRL = systemIcon("ErasingIcon")
	ICON_SWITCH = systemIcon("General")
	ICON_SYNC = systemIcon("Sync")
	ICON_TRASH = systemIcon("TrashIcon")
	ICON_USER = systemIcon("UserIcon")
	ICON_WARNING = systemIcon("AlertCautionIcon")
	ICON_WEB = systemIcon("BookmarkIcon")
}
