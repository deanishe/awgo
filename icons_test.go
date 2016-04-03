package workflow

import (
	"os"
	"testing"
)

func TestIcons(t *testing.T) {
	icons := []*ItemIcon{
		ICON_ACCOUNT,
		ICON_BURN,
		ICON_CLOCK,
		ICON_COLOR,
		ICON_COLOUR,
		ICON_EJECT,
		ICON_ERROR,
		ICON_FAVORITE,
		ICON_FAVOURITE,
		ICON_GROUP,
		ICON_HELP,
		ICON_HOME,
		ICON_INFO,
		ICON_NETWORK,
		ICON_NOTE,
		ICON_SETTINGS,
		ICON_SWIRL,
		ICON_SWITCH,
		ICON_SYNC,
		ICON_TRASH,
		ICON_USER,
		ICON_WARNING,
		ICON_WEB,
	}
	for _, icon := range icons {
		if icon.Type != "" {
			t.Fatalf("icon.Type is not empty: %v", icon.Value)
		}
		_, err := os.Stat(icon.Value)
		if err != nil {
			t.Fatalf("Couldn't stat %v: %v", icon.Value, err)
		}

	}
}
