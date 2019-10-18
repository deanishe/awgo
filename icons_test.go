// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIcons(t *testing.T) {
	t.Parallel()

	icons := []*Icon{
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
		icon := icon
		t.Run(icon.Value, func(t *testing.T) {
			assert.Equal(t, IconType(""), icon.Type, "icon.Type is not empty")

			// Skip path validation on Travis because it's a Linux box
			if os.Getenv("TRAVIS") != "" {
				return
			}
			_, err := os.Stat(icon.Value)
			assert.Nil(t, err, "stat failed")
		})
	}
}
