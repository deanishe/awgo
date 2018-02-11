//
// Copyright (c) 2016 Dean Jackson <deanishe@deanishe.net>
//
// MIT Licence. See http://opensource.org/licenses/MIT
//

package aw

// IconType specifies the type of an aw.Icon struct. It can be an image file,
// the icon of a file, e.g. an application's icon, or the icon for a UTI.
type IconType string

// Valid icon types.
const (
	// Indicates that Icon.Value is the path to an image file that should
	// be used as the Item's icon.
	IconTypeImage IconType = ""
	// Icon.Value points to an object whose icon should be show in Alfred,
	//e.g. combine with "/Applications/Safari.app" to show Safari's icon.
	IconTypeFileIcon IconType = "fileicon"
	// Indicates that Icon.Value is a UTI, e.g. "public.folder",
	// which will give you the icon for a folder.
	IconTypeFileType IconType = "filetype"
)

// Ready-to-use icons based on macOS system icons. These icons are all found in
//
//     /System/Library/CoreServices/CoreTypes.bundle/Contents/Resources
//
// The icons are the same as found in the Alfred-Workflow library
// for Python. Preview them here:
// http://www.deanishe.net/alfred-workflow/user-manual/icons.html#list-of-icons
var (
	// Workflow's own icon
	IconWorkflow = &Icon{"icon.png", IconTypeImage}

	sysIcons = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/"
	// System icons
	IconAccount   = &Icon{Value: sysIcons + "Accounts.icns"}
	IconBurn      = &Icon{Value: sysIcons + "BurningIcon.icns"}
	IconClock     = &Icon{Value: sysIcons + "Clock.icns"}
	IconColor     = &Icon{Value: sysIcons + "ProfileBackgroundColor.icns"}
	IconColour    = &Icon{Value: sysIcons + "ProfileBackgroundColor.icns"}
	IconEject     = &Icon{Value: sysIcons + "EjectMediaIcon.icns"}
	IconError     = &Icon{Value: sysIcons + "AlertStopIcon.icns"}
	IconFavorite  = &Icon{Value: sysIcons + "ToolbarFavoritesIcon.icns"}
	IconFavourite = &Icon{Value: sysIcons + "ToolbarFavoritesIcon.icns"}
	IconGroup     = &Icon{Value: sysIcons + "GroupIcon.icns"}
	IconHelp      = &Icon{Value: sysIcons + "HelpIcon.icns"}
	IconHome      = &Icon{Value: sysIcons + "HomeFolderIcon.icns"}
	IconInfo      = &Icon{Value: sysIcons + "ToolbarInfo.icns"}
	IconNetwork   = &Icon{Value: sysIcons + "GenericNetworkIcon.icns"}
	IconNote      = &Icon{Value: sysIcons + "AlertNoteIcon.icns"}
	IconSettings  = &Icon{Value: sysIcons + "ToolbarAdvanced.icns"}
	IconSwirl     = &Icon{Value: sysIcons + "ErasingIcon.icns"}
	IconSwitch    = &Icon{Value: sysIcons + "General.icns"}
	IconSync      = &Icon{Value: sysIcons + "Sync.icns"}
	IconTrash     = &Icon{Value: sysIcons + "TrashIcon.icns"}
	IconUser      = &Icon{Value: sysIcons + "UserIcon.icns"}
	IconWarning   = &Icon{Value: sysIcons + "AlertCautionIcon.icns"}
	IconWeb       = &Icon{Value: sysIcons + "BookmarkIcon.icns"}
)

// Icon represents the icon for an Item.
//
// Alfred can show icons based on image files, UTIs (e.g. "public.folder") or
// can use the icon of a specific file (e.g. "/Applications/Safari.app"
// to use Safari's icon.
//
// Type = "" (the default) will treat Value as the path to an image file.
// Alfred supports (at least) PNG, ICNS, JPG, GIF.
//
// Type = IconTypeFileIcon will treat Value as the path to a file or
// directory and use that file's icon, e.g:
//
//    icon := &Icon{"/Applications/Mail.app", IconTypeFileIcon}
//
// will display Mail.app's icon.
//
// Type = IconTypeFileType will treat Value as a UTI, such as
// "public.movie" or "com.microsoft.word.doc". UTIs are useful when
// you don't have a local path to point to.
//
// You can find out the UTI of a filetype by dragging one of the files
// to a File Filter's File Types list in Alfred, or in a shell with:
//
//    mdls -name kMDItemContentType -raw /path/to/the/file
//
// This will only work on Spotlight-indexed files.
type Icon struct {
	Value string   `json:"path"`           // Path or UTI
	Type  IconType `json:"type,omitempty"` // "fileicon", "filetype" or ""
}
