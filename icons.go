//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

package aw

import "fmt"

// Valid icon types for Icon. You can use an image file, the icon of a file,
// e.g. an application's icon, or the icon for a filetype (specified by a UTI).
const (
	// For image files you wish to show in Alfred.
	IconTypeImageFile = ""
	// Show the icon of a file, e.g. combine with "/Applications/Safari.app"
	// to show Safari's icon in Alfred.
	IconTypeFileIcon = "fileicon"
	// For UTIs to show the icon for a filetype, e.g. "public.folder",
	// which will give you the icon for a folder.
	IconTypeFileType = "filetype"
)

// Ready-to-use icons based on built-in OS X system icons.
// These icons are all found in
// /System/Library/CoreServices/CoreTypes.bundle/Contents/Resources.
//
// The icons are the same as found in the Alfred-Workflow library
// for Python. Preview them here:
// http://www.deanishe.net/alfred-workflow/user-manual/icons.html#list-of-icons
var (
	IconAccount   *Icon // Accounts.icns
	IconBurn      *Icon // BurningIcon.icns
	IconClock     *Icon // Clock.icns
	IconColor     *Icon // ProfileBackgroundColor.icns
	IconColour    *Icon // ProfileBackgroundColor.icns
	IconEject     *Icon // EjectMediaIcon.icns
	IconError     *Icon // AlertStopIcon.icns
	IconFavorite  *Icon // ToolbarFavoritesIcon.icns
	IconFavourite *Icon // ToolbarFavoritesIcon.icns
	IconGroup     *Icon // GroupIcon.icns
	IconHelp      *Icon // HelpIcon.icns
	IconHome      *Icon // HomeFolderIcon.icns
	IconInfo      *Icon // ToolbarInfo.icns
	IconNetwork   *Icon // GenericNetworkIcon.icns
	IconNote      *Icon // AlertNoteIcon.icns
	IconSettings  *Icon // ToolbarAdvanced.icns
	IconSwirl     *Icon // ErasingIcon.icns
	IconSwitch    *Icon // General.icns
	IconSync      *Icon // Sync.icns
	IconTrash     *Icon // TrashIcon.icns
	IconUser      *Icon // UserIcon.icns
	IconWarning   *Icon // AlertCautionIcon.icns
	IconWeb       *Icon // BookmarkIcon.icns
)

// Icon represents the icon for an Item.
//
// Alfred supports PNG or ICNS files, UTIs (e.g. "public.folder") or
// can use the icon of a specific file (e.g. "/Applications/Safari.app"
// to use Safari's icon.
//
// Type = "" (the default) will treat Value as the path to a PNG or ICNS
// file.
//
// Type = "fileicon" will treat Value as the path to a file or directory
// and use that file's icon, e.g:
//
//    icon := Icon{"/Applications/Mail.app", "fileicon"}
//
// will display Mail.app's icon.
//
// Type = "filetype" will treat Value as a UTI, such as "public.movie"
// or "com.microsoft.word.doc". UTIs are useful when you don't have
// a local path to point to.
//
// You can find out the UTI of a filetype by dragging one of the files
// to a File Filter's File Types list in Alfred, or in a shell with:
//
//    mdls -name kMDItemContentType -raw /path/to/the/file
//
// This will only work on Spotlight-indexed files.
type Icon struct {
	Value string `json:"path"`
	Type  string `json:"type,omitempty"`
}

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
