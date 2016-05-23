package workflow

import (
	"os"
	"testing"
)

func TestIcons(t *testing.T) {
	icons := []*ItemIcon{
		IconAccount,
		IconBurn,
		IconClock,
		IconColor,
		IconColour,
		IconEject,
		IconError,
		IconFavorite,
		IconFavourite,
		IconGroup,
		IconHelp,
		IconHome,
		IconInfo,
		IconNetwork,
		IconNote,
		IconSettings,
		IconSwirl,
		IconSwitch,
		IconSync,
		IconTrash,
		IconUser,
		IconWarning,
		IconWeb,
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
